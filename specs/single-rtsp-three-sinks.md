# Single RTSP, three sinks

## Status

Draft. Awaiting Phase 0 (ffmpeg validation gate) and user sign-off on the three open decisions before any Go code changes.

## Problem

`corvette-core` currently opens two independent RTSP sessions per camera. One feeds the segmented-MP4 recorder. The other feeds the AI inference pipe at 2 fps rawvideo. A planned live-view UI would force a third session, doubling the camera's connection load and the ffmpeg decoder cost for every camera the NVR runs.

## Goal

One RTSP session per camera. One ffmpeg process. Three output sinks fed from that single demux+decode.

- **Recorder.** Segmented MP4, `-c copy`, no frame loss, no backpressure from the other sinks.
- **AI inference.** 640x640, 2 fps, raw rgb24 piped to Go. Drops frames on overflow instead of stalling the recorder.
- **Live stream.** Reserved third `-map` slot. Not built in this change, but the interface and process shape leave room for it.

## Non-goals

- Implementing the live stream itself.
- Changing recording file layout, retention, or playback.
- Changing AI inference logic, model, or frame format.
- Changing restart behavior, camera registry, vendor model, or config schema beyond one new field.

## Open decisions

These three forks are the user's call. None of them can be settled by running a probe.

1. **ffmpeg-go chain form.** Chained `.Input().Output().Output().Compile()` is shorter and maps 1:1 to the CLI. `OutputContext` with `[]*Stream` is more debuggable. Default: chained.
2. **Future live sink format.** MJPEG over HTTP (low CPU, ~200ms latency, simple `<img>` tag) or HLS segments (browser-native, higher CPU, ~2-5s latency). Default for the home-NUC / 4-5 camera target: MJPEG.
3. **AI reader non-blocking send.** The `select { default: discard }` pattern before `aiFrameChan <- frame` is the last line of defense against backpressure escaping the fifo and reaching the recorder. Default: include it.

## Architecture (one diagram, the rest of the spec points back here)

```
RTSP source
   |
   v
[1 RTSP demux + 1 decode]
   |
   |--- -map 0:v -map 0:a -c copy
   |       -f fifo (block on overflow, default)
   |       -f segment -segment_time 5 ...
   |       -> ./recordings/<id>/rec_%Y-%m-%d_%H-%M-%S.mp4
   |
   |--- -map 0:v -vf 'fps=2,scale=640:640' -pix_fmt rgb24
   |       -f fifo -drop_pkts_on_overflow 1 -queue_size 30
   |       -f rawvideo
   |       -> pipe:1 (Go AI reader)
   |
   `--- (future) -map 0:v -c:v mjpeg
           -f fifo -drop_pkts_on_overflow 1
           -f mjpeg
           -> pipe:1 (Go HTTP MJPEG endpoint)
```

The `fifo` muxer is the load-bearing piece. Each sink runs in its own ffmpeg-internal thread with its own queue. The recorder fifo uses default policy (block on full, never drop). The AI fifo flips `drop_pkts_on_overflow=1` so a slow Go reader can lose frames without dragging the recorder down with it. The future live fifo will use the same drop policy as the AI sink.

## Risks tracked (from blast-radius review)

| Risk                                                                                                     | Severity | Where it shows up                                                                                                                            |
| -------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Demuxer `fflags nobuffer` / `flags low_delay` from old AI args leaks into recorder and corrupts segments | High     | Phase 0 (validate the merged command), Phase 1 (interface does not touch demuxer args)                                                       |
| `fifo` wrapping the segment muxer behaves unexpectedly with `-c copy`                                    | High     | Phase 0                                                                                                                                      |
| `-map 0:a` fails on no-audio cameras                                                                     | Medium   | Phase 0 (test on a no-audio source if available), Phase 2 (drop the explicit `0:a` map if `-c copy` carries the source's streams on its own) |
| `aiFrameChan` lifecycle race on restart                                                                  | Medium   | Phase 2 (channel moves from constructor into `Start`)                                                                                        |
| `ffmpeg-go` chain silently emits two processes                                                           | High     | Phase 0 (`ps -ef` check), Phase 2 (`ps -ef` re-check)                                                                                        |

---

# Phases

Each phase is self-contained. The ffmpeg command, the code shape, and the verification it owns all live inside the phase. Stop after any phase and the work is resumable.

## Phase 0. Validate the ffmpeg command

**Goal.** Prove the reference ffmpeg command behaves correctly before any Go code moves. The only phase with no Go changes.

**The command to run.** Substitute your camera's URL. Example for a local 192.168.100.74 camera with no auth:

````bash
URL="rtsp://192.168.100.74:554/live/ch00_1"
OUT="./recordings/test"
mkdir -p "$OUT"

```bash
ffmpeg -rtsp_transport tcp -timeout 5000000 -i "$URL" \
  -map 0:v -c copy \
    -f fifo -attempt_recovery 1 \
    -f segment -segment_time 5 -reset_timestamps 1 -strftime 1 \
    -segment_format_options movflags=frag_keyframe+empty_moov+default_base_moof \
    "$OUT/rec_%Y-%m-%d_%H-%M-%S.mp4" \
  -map 0:v -vf 'fps=2,scale=640:640' -pix_fmt rgb24 -f rawvideo \
    -f fifo -queue_size 30 -drop_pkts_on_overflow 1 \
    pipe:1 | wc -c
````

**If your camera needs credentials, swap the URL line for `URL="rtsp://user:pass@192.168.100.74:554/live/ch00_1"`.**

**Audio is optional.** The reference command includes `-map 0:a` because some cameras expose audio. If yours doesn't, ffmpeg fails with `Stream map '' matches no streams`. Drop the audio map line and re-run. The video-only command is:

```bash
URL="rtsp://192.168.100.74:554/live/ch00_1"
OUT="./recordings/test"
mkdir -p "$OUT"

ffmpeg -rtsp_transport tcp -timeout 5000000 -i "$URL" \
  -map 0:v -c copy \
    -f fifo -attempt_recovery 1 \
    -f segment -segment_time 5 -reset_timestamps 1 -strftime 1 \
    -segment_format_options movflags=frag_keyframe+empty_moov+default_base_moof \
    "$OUT/rec_%Y-%m-%d_%H-%M-%S.mp4" \
  -map 0:v -vf 'fps=2,scale=640:640' -pix_fmt rgb24 -f rawvideo \
    -f fifo -queue_size 30 -drop_pkts_on_overflow 1 \
    pipe:1 | wc -c
```

**Steps.**

1. Pick any reachable RTSP source (real camera, public test stream, or local `testsrc` loop).
2. Run the command above in one terminal. In another terminal, leave `wc -c` running on the pipe (it's already in the command).
3. After ~60 seconds, run the four checks below.
4. Run the backpressure test: stop the pipe consumer (`Ctrl-C` on the `wc -c` consumer, or close its stdin), wait 15 seconds, re-run the file checks.

**Four checks, all in one shell:**

```bash
# 1. Process count: must be 1, not 2
ps -ef | grep '[f]fmpeg' | wc -l

# 2. File timing: one MP4 every ~5s, durations ~5s
ls -lt "$OUT" | head -20
for f in $(ls -t "$OUT" | head -3); do
  ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/$f"
done

# 3. Pipe throughput: ~2.4 MB/s on a stream with motion
# wc -c should be printing a number that grows by ~2.4MB per second

# 4. Backpressure safety: after killing wc -c, MP4s keep arriving
ls -lt "$OUT" | head -5
ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/$(ls -t "$OUT" | head -1)"
```

**Done when.** All four checks pass. Process count is 1. MP4s land every ~5s and are `ffprobe`-clean. Pipe shows ~2.4 MB/s. After killing the pipe consumer, MP4s keep landing and stay clean.

**If a check fails, fix the ffmpeg command, not the Go code.** Re-run until clean.

**Exit artifact.** A short log (or terminal capture) of the four checks, attached to the PR description or pasted in a comment.

## Phase 1. Update the Streamer interface

**Goal.** Reshape the contract so Phase 2 has a stable target. No behavior change yet.

**The new shape, in full:**

In `internal/streamer/streamer.go`:

```go
type Streamer interface {
    Start(eGCtx context.Context) error
    Stop() error
    GetAIFrame() chan []float32
    Vendor() vendors.Vendor
    GetPath() string
    Destroy()
}

type CreateRtspStreamerOpts struct {
    RtspVendor   vendors.Vendor
    ScalingSize  int
    StreamingFps int
    Outputs      []OutputKind
}

type OutputKind int

const (
    OutputRecording OutputKind = iota
    OutputAIInference
    OutputLive // future
)
```

`StreamerMapper` passes `Outputs: []OutputKind{OutputRecording, OutputAIInference}`.

**Steps.**

1. In `internal/streamer/streamer.go`, replace the `Streamer` interface with the shape above.
2. Add `OutputKind` and the `Outputs` field to `CreateRtspStreamerOpts`.
3. Update `StreamerMapper` to pass `Outputs: []OutputKind{OutputRecording, OutputAIInference}`.
4. Run `go build ./...`. Expect compile errors in `internal/streamer/rtsp.go` because `RtspStreamer` no longer satisfies the new interface. **This is expected.** Do not fix `rtsp.go` in this phase.
5. Comment out or stub the call sites in `internal/camera/handler.go` that use the old method names so the project builds. Stubs can return `nil` or panic; the point is to get the project compiling against the new interface shape.

**Done when.** `go build ./...` succeeds. `RtspStreamer` either satisfies the new interface or is the only remaining compile error, and the stub at the call site compiles.

**Exit artifact.** Diff in `streamer.go` and the stub at the handler call site.

## Phase 2. Rewrite RtspStreamer

**Goal.** One ffmpeg process, two `-map` outputs, the AI reader goroutine reading from the shared pipe. The actual migration.

**The new struct, in full:**

In `internal/streamer/rtsp.go`:

```go
type RtspStreamer struct {
    rtspVendor   vendors.Vendor
    mu           sync.Mutex
    cmd          *exec.Cmd
    aiFrameChan  chan []float32
    inputBuf     []float32
    scalingSize  int
    streamingFps int
    // future: liveFrameChan chan []byte
}

func CreateRtspStreamer(opts *CreateRtspStreamerOpts) *RtspStreamer
func (rs *RtspStreamer) Start(eGCtx context.Context) error
func (rs *RtspStreamer) Stop() error
func (rs *RtspStreamer) readAiFrames(stdout io.ReadCloser, ctx context.Context) error
func (rs *RtspStreamer) GetAIFrame() chan []float32
func (rs *RtspStreamer) Vendor() vendors.Vendor
func (rs *RtspStreamer) GetPath() string
func (rs *RtspStreamer) Destroy()
```

`aiFrameChan` moves from `CreateRtspStreamer` into `Start`. Channel lifetime now matches the process. `Destroy` becomes a no-op or just marks the streamer torn down. The registry's restart loop calls `Start` on a clean streamer, never on a destroyed one, so the old send-to-closed-channel race disappears.

**The new handler, in full:**

In `internal/camera/handler.go`, `StartAllFunctions` loses one of its three `g.Go` lines:

```go
g.Go(func() error { return ch.streamer.Start(ctx) })
g.Go(func() error { return ch.StartAIInference(ctx) })
// future: g.Go(func() error { return ch.streamer.ConsumeLive(ctx) })
```

**Steps.**

1. Rewrite `RtspStreamer` in `internal/streamer/rtsp.go` to the shape above. One `cmd`, one `Start`, one `Stop`.
2. Build the ffmpeg command with `ffmpeg_go.Input(url, inputArgs).Output(recArgs, recPath).Output(aiArgs, "pipe:1").Compile()`. Mirror the Phase 0 reference command exactly. If Phase 0 passed, this is a mechanical translation.
3. Move `aiFrameChan` creation from `CreateRtspStreamer` into `Start`. Channel lifetime now matches the process.
4. Implement the AI reader goroutine: `readAiFrames` blocks on `io.ReadFull`, sends to `aiFrameChan` with a non-blocking `select { case rs.aiFrameChan <- frame: default: }`.
5. Implement `Stop` to kill the single `*exec.Cmd`. No second command to track.
6. Remove the old `StartRecording`, `StartAIStreaming`, `StopRecording`, `StopAIStreaming`, `cleanupRecorder`, `cleanupAiStreamer`, `waitForCommand` methods. They have no replacement.
7. Update `internal/camera/handler.go` to call `ch.streamer.Start(ctx)` once and remove the stub from Phase 1.
8. Run `go vet ./...`, `go build ./...`.

**Done when.** Clean build, no `TODO` comments left in the diff, every removed method has a replacement or is provably dead.

**Exit artifact.** Diff in `rtsp.go` and `handler.go`.

## Phase 3. Verify

**Goal.** Prove the migration works against real cameras at the scale you actually run.

**Steps.**

1. Run with one camera for 60 seconds. Confirm: recordings land every ~5s, `ffprobe` shows clean MP4s, AI inference runs (log line or test), ffmpeg process count is exactly 1 for that camera.
2. `SIGTERM` corvette. Confirm clean shutdown: ffmpeg process exits, no zombie, no panic in logs, camera service marks camera offline.
3. Restart corvette. Confirm the registry's restart loop brings the camera back online without operator action.
4. Run with two cameras concurrently. Confirm: ffmpeg process count is 2 (not 4), no shared-state corruption between the two streamers, both record cleanly.
5. If you have audio on the test cameras, confirm audio plays back in the recordings. If not, confirm the no-audio path doesn't fail.
6. Run the AI pipe stall test: artificially stall `DoInference` for 10 seconds, confirm recordings keep landing during the stall, AI frames resume after.

**Done when.** All six checks pass on real hardware. The previous behavior (two-process shape) is observably worse on at least one of: ffmpeg process count, ffmpeg memory, network sockets to the camera.

**Exit artifact.** Terminal capture of the four run-time checks, plus a `ps -ef | grep ffmpeg` snapshot showing one process per camera.

## Phase 4. Add the live stream sink (future, not this PR)

**Goal.** Drop a third `-map` output into the existing single-process shape. The migration to one process in Phases 0-3 is what makes this cheap.

**Steps.**

1. Decide MJPEG vs HLS. Default for the home-NUC / 4-5 camera target is MJPEG.
2. Add `OutputLive` to the active `Outputs` list in `StreamerMapper`.
3. Add `liveFrameChan chan []byte` to `RtspStreamer`. Add `GetLiveFrame() chan []byte` to the `Streamer` interface.
4. Add the third `-map` output in the `Start` body: `-c:v mjpeg -q:v 5 -f fifo -drop_pkts_on_overflow 1 -f mjpeg pipe:1`.
5. Add a reader goroutine in `Start` that frames MJPEG and pushes to `liveFrameChan`.
6. Add a `g.Go(func() error { return ch.streamer.ConsumeLive(ctx) })` line in `CameraHandler.StartAllFunctions` that reads from `liveFrameChan` and writes to a registered HTTP handler.
7. Add an HTTP handler in `internal/platform/handler/handlers/` (or wherever the existing camera routes live) that exposes `/live/<id>.mjpeg` as `multipart/x-mixed-replace`.
8. Re-run the Phase 3 verification matrix with the live sink active.

**Done when.** Browser can open `http://<nuc>:9090/live/1.mjpeg` and show a live image. Recordings still work. AI still work. ffmpeg process count is still 1 per camera.

**Exit artifact.** Diff adding the live stream, plus a screenshot or screen recording of the browser view.

---

## Phases at a glance

| Phase                   | What it ships                             | Done when                                        | Reversible                    |
| ----------------------- | ----------------------------------------- | ------------------------------------------------ | ----------------------------- |
| 0. Validate ffmpeg      | Reference command proven                  | All four validation checks pass                  | Yes, no code touched          |
| 1. Update interface     | New `Streamer` shape, stubs at call sites | `go build` clean                                 | Yes, easy revert              |
| 2. Rewrite RtspStreamer | One-process migration                     | `go vet` + `go build` clean                      | Yes, branch reverts cleanly   |
| 3. Verify               | Runtime proof on real cameras             | All six checks pass                              | Yes, can roll back deployment |
| 4. Live stream (future) | Third `-map` output + HTTP endpoint       | Browser shows live image, prior gates still pass | Yes, additive change          |

## Unchanged across all phases

- `internal/camera/registry.go` (restart loop, handler map, watchdog wiring).
- `cmd/main.go` (wiring through `StreamerMapper` to `RegisterArrStreamers`).
- `internal/vendors/` (pure getters on `RtspVendor`).
- `internal/services/`, `internal/database/`, `internal/playback/`, `internal/object_detection/`, `internal/config/`.
