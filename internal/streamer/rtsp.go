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
	RtspVendor *vendors.RtspVendor

	ScalingSize  int
	StreamingFps int
}

type RtspStreamer struct {
	rtspVendor *vendors.RtspVendor
	mu         sync.Mutex

	// Recording
	isRecording StreamState
	recCmd      *exec.Cmd

	// AIStream
	scalingSize  int
	isStreaming  StreamState
	aiCmd        *exec.Cmd
	aiFrameChan  chan []byte
	streamingFps int
}

func CreateRtspStreamer(opts *CreateRtspStreamerOpts) *RtspStreamer {
	if DoesThisFolderExist(opts.RtspVendor.CamName) == false {
		CreateFolder(opts.RtspVendor.CamName)
	}

	return &RtspStreamer{
		rtspVendor:  opts.RtspVendor,
		isRecording: Stopped,

		scalingSize:  opts.ScalingSize,
		isStreaming:  Stopped,
		aiFrameChan:  make(chan []byte, 1),
		streamingFps: opts.StreamingFps,
	}
}

func (rs *RtspStreamer) StartRecording(eGCtx context.Context) error {
	rs.mu.Lock()

	if rs.isRecording == Running {
		rs.mu.Unlock()
		slog.Warn("Start recording called whilst recorder is already started.", "for", rs.rtspVendor.CamName)
		return ErrAlreadyRecording
	}

	folderPath := getPath(rs.rtspVendor.CamName)
	dirPath := filepath.Join(folderPath, `out_%d-%m-%Y-%H-%M-%S.mp4`)

	inputArgs := ffmpeg_go.KwArgs{
		"rtsp_transport": "tcp",
		"timeout":        "5000000",
	}

	outputArgs := ffmpeg_go.KwArgs{
		"c":                "copy",
		"f":                "segment",
		"segment_time":     "300",
		"reset_timestamps": "1",
		"strftime":         "1",
	}

	template := ffmpeg_go.Input(rs.rtspVendor.URL, inputArgs)
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

	return rs.waitForCommand(command, &errBuff)
}

func (rs *RtspStreamer) waitForCommand(command *exec.Cmd, errBuf *bytes.Buffer) error {
	defer rs.cleanupRecorder(command)
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
	slog.Info("Stop recording function tirggered.", "for", rs.rtspVendor.CamName)

	if err := rs.recCmd.Process.Kill(); err != nil {
		slog.Error("Failed to kill process.", "for", rs.rtspVendor.CamName, "err", err.Error())
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
		"rtsp_transport": "tcp",
		"timeout":        "5000000",
	}

	outputArgs := ffmpeg_go.KwArgs{
		"vf":      fmt.Sprintf("fps=%d,scale=640:640", rs.streamingFps),
		"c:v":     "rawvideo",
		"pix_fmt": "rgb24",
		"f":       "rawvideo",
	}

	template := ffmpeg_go.Input(rs.rtspVendor.URL, inputArgs)
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
				return err
			}
		}
		select {
		case <-ctx.Done():
			slog.Info("Other function failed, quitting AI streamer.")
			return ctx.Err()
		case rs.aiFrameChan <- frameBuffer:
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
	close(rs.aiFrameChan)
	slog.Info("AI Streamer cleanedup.")
}

func (rs *RtspStreamer) StopAIStreaming() error {
	rs.mu.Lock()

	if rs.aiCmd == nil || rs.aiCmd.Process == nil {
		return nil
	}
	if err := rs.aiCmd.Process.Kill(); err != nil {
		slog.Error("Failed to kill process.", "for", rs.rtspVendor.CamName, "err", err.Error())
		return err
	}
	return nil
}

func (rs *RtspStreamer) GetAIFrame() []byte {
	select {
	case data := <-rs.aiFrameChan:
		return data
	}
}
