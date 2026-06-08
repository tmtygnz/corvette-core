package streamer

import (
	"bytes"
	"context"
	"corvette/internal/vendors"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"sync"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

var ErrAlreadyRecording = errors.New("Start recording called whilst already recording.")

var ErrAlreadyStreaming = errors.New("Start AI streaming called whilst already streaming.")

type StreamState int

const (
	Stopped = iota
	Running
)

type CreateRtspStreamerOpts struct {
	RtspVendor vendors.Vendor

	ScalingSize  int
	StreamingFps int
}

type RtspStreamer struct {
	rtspVendor vendors.Vendor
	mu         sync.Mutex

	// Recording
	isRecording StreamState
	recCmd      *exec.Cmd

	// AIStream
	scalingSize  int
	isStreaming  StreamState
	aiCmd        *exec.Cmd
	aiFrameChan  chan []float32
	inputBuf     []float32
	streamingFps int
}

func CreateRtspStreamer(opts *CreateRtspStreamerOpts) *RtspStreamer {
	if DoesThisFolderExist(opts.RtspVendor.CamName()) == false {
		CreateFolder(opts.RtspVendor.CamName())
	}

	return &RtspStreamer{
		rtspVendor:  opts.RtspVendor,
		isRecording: Stopped,

		scalingSize:  opts.ScalingSize,
		isStreaming:  Stopped,
		aiFrameChan:  make(chan []float32, 1),
		streamingFps: opts.StreamingFps,
		inputBuf:     make([]float32, opts.ScalingSize*opts.ScalingSize*3),
	}
}

func (rs *RtspStreamer) StartRecording(eGCtx context.Context) error {
	rs.mu.Lock()

	if rs.isRecording == Running {
		rs.mu.Unlock()
		slog.Warn("Start recording called whilst recorder is already started.", "for", rs.rtspVendor.CamName())
		return ErrAlreadyRecording
	}

	folderPath := getPath(rs.rtspVendor.CamName())
	dirPath := filepath.Join(folderPath, `out_%d-%m-%Y-%H-%M-%S.mp4`)

	inputArgs := ffmpeg_go.KwArgs{
		"rtsp_transport":  "tcp",
		"timeout":         "5000000",
		"analyzeduration": "2000000",
	}

	outputArgs := ffmpeg_go.KwArgs{
		"c":                      "copy",
		"f":                      "segment",
		"segment_time":           "300",
		"reset_timestamps":       "1",
		"strftime":               "1",
		"segment_format_options": "movflags=frag_keyframe+empty_moov+default_base_moof",
	}

	template := ffmpeg_go.Input(rs.rtspVendor.URL(), inputArgs)
	templateWithContext := ffmpeg_go.OutputContext(eGCtx, []*ffmpeg_go.Stream{template}, dirPath, outputArgs).GlobalArgs("-loglevel", "error")

	command := templateWithContext.Compile()
	var errBuff bytes.Buffer
	command.Stderr = &errBuff

	rs.isRecording = Running
	rs.recCmd = command

	if err := command.Start(); err != nil {
		rs.isRecording = Stopped
		rs.recCmd = nil
		rs.mu.Unlock()
		slog.Error("Failed to start recording command.", "err", err.Error())
		return err
	}
	rs.mu.Unlock()

	return rs.waitForCommand(eGCtx, command, &errBuff)
}

func (rs *RtspStreamer) waitForCommand(ctx context.Context, command *exec.Cmd, errBuf *bytes.Buffer) error {
	defer rs.cleanupRecorder(command)
	go func() {
		<-ctx.Done()
		if command.Process != nil {
			command.Process.Kill()
		}
	}()
	err := command.Wait()
	if err != nil {
		slog.Error("ffmpeg command returned.", "err", err.Error())
		return err
	}
	if errBuf.Len() > 0 {
		slog.Error("ffmpeg returned", "err", errBuf.String())
		return errors.New("ffmpeg unpredictable error.")
	}

	return nil
}

func (rs *RtspStreamer) cleanupRecorder(cmd *exec.Cmd) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.recCmd != cmd {
		return
	}

	rs.recCmd = nil
	rs.isRecording = Stopped
	slog.Info("Recorder cleaned up.")
}

func (rs *RtspStreamer) StopRecording() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.recCmd == nil || rs.recCmd.Process == nil {
		return nil
	}
	slog.Info("Stop recording function tirggered.", "for", rs.rtspVendor.CamName())

	if err := rs.recCmd.Process.Kill(); err != nil {
		slog.Error("Failed to kill process.", "for", rs.rtspVendor.CamName(), "err", err.Error())
		return err
	}
	return nil
}

func (rs *RtspStreamer) StartAIStreaming(eGCtx context.Context) error {
	rs.mu.Lock()

	if rs.isStreaming == Running {
		rs.mu.Unlock()
		slog.Warn("Start AI stream called even though it is already running.")
		return ErrAlreadyRecording
	}

	inputArgs := ffmpeg_go.KwArgs{
		"hwaccel":        "qsv",
		"rtsp_transport": "tcp",
		"timeout":        "5000000",
		"fflags":         "nobuffer",
		"flags":          "low_delay",
	}

	outputArgs := ffmpeg_go.KwArgs{
		"vf":      fmt.Sprintf("fps=%d,vpp_qsv=w=640:h=640", rs.streamingFps),
		"c:v":     "rawvideo",
		"pix_fmt": "rgb24",
		"f":       "rawvideo",
		"threads": "1",
	}

	var url string
	if rs.rtspVendor.SURL() != "" {
		url = rs.rtspVendor.SURL()
	} else {
		url = rs.rtspVendor.URL()
	}

	template := ffmpeg_go.Input(url, inputArgs)
	templateWithContext := ffmpeg_go.OutputContext(eGCtx, []*ffmpeg_go.Stream{template}, "pipe:1", outputArgs).GlobalArgs("-loglevel", "error")

	command := templateWithContext.Compile()

	output, err := command.StdoutPipe()
	if err != nil {
		rs.mu.Unlock()
		slog.Error("Failed to get output buffer.", "err", err.Error())
		return err
	}

	var stderr bytes.Buffer
	command.Stderr = &stderr

	rs.aiCmd = command
	rs.isStreaming = Running

	if err := command.Start(); err != nil {
		rs.isStreaming = Stopped
		rs.aiCmd = nil
		rs.mu.Unlock()
		slog.Error("Failed to start ai streaming command.", "err", err.Error())
		return err
	}
	rs.mu.Unlock()

	return rs.readAiFrames(output, eGCtx, command)
}

func (rs *RtspStreamer) readAiFrames(stdout io.ReadCloser, ctx context.Context, cmd *exec.Cmd) error {
	defer rs.cleanupAiStreamer(cmd)

	go func() {
		<-ctx.Done()
		stdout.Close()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	frameBuffer := make([]byte, rs.scalingSize*rs.scalingSize*3) // w*h * 3 colors (r,g,b)
	for {
		_, err := io.ReadFull(stdout, frameBuffer)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				slog.Error("Failed to read stdout of AI stream.", "err", err.Error())
				return fmt.Errorf("AI Stream lost.")
			}
		}
		select {
		case <-ctx.Done():
			slog.Info("Other function failed, quitting AI streamer.")
			return ctx.Err()
		case rs.aiFrameChan <- rs.prepareInput(frameBuffer):
		}
	}
}

func (rs *RtspStreamer) cleanupAiStreamer(cmd *exec.Cmd) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.aiCmd != cmd {
		return
	}

	rs.aiCmd = nil
	rs.isStreaming = Stopped
	slog.Info("AI Streamer cleanedup.")
}

func (rs *RtspStreamer) prepareInput(frame []byte) []float32 {
	size := rs.scalingSize * rs.scalingSize

	for i := range size {
		// RGB Interleaved (byte) -> RGB Planar (float32)
		rs.inputBuf[i] = float32(frame[i*3]) / 255.0          // R
		rs.inputBuf[i+size] = float32(frame[i*3+1]) / 255.0   // G
		rs.inputBuf[i+size*2] = float32(frame[i*3+2]) / 255.0 // B
	}

	return rs.inputBuf
}

func (rs *RtspStreamer) StopAIStreaming() error {
	rs.mu.Lock()

	if rs.aiCmd == nil || rs.aiCmd.Process == nil {
		return nil
	}
	if err := rs.aiCmd.Process.Kill(); err != nil {
		slog.Error("Failed to kill process.", "for", rs.rtspVendor.CamName(), "err", err.Error())
		return err
	}
	return nil
}

func (rs *RtspStreamer) GetAIFrame() chan []float32 {
	return rs.aiFrameChan
}

func (rs *RtspStreamer) Vendor() vendors.Vendor {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.rtspVendor
}

func (rs *RtspStreamer) Destroy() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	close(rs.aiFrameChan)
}
