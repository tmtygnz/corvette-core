package capture

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type CreateNewGenericCapturerOpts struct {
	URL   string
	Name  string
	AiFPS int
}

type RtspCapturer struct {
	URL   string
	Name  string
	Aifps int

	recCmd *exec.Cmd
	aiCmd  *exec.Cmd
}

// TODO: Turn parameters into a struct.
func CreateNewCapturer(opts CreateNewGenericCapturerOpts) *RtspCapturer {
	log.Printf("New recorder for: %s at %s", opts.Name, opts.URL)
	setupFolderForCapturer(opts.Name)
	return &RtspCapturer{
		opts.URL,
		opts.Name,
		2,
		nil,
		nil,
	}
}

func (rfg *RtspCapturer) StartRecorder() error {
	log.Printf("Camera streaming started for: %s\n", rfg.Name)
	dirPath := fmt.Sprintf("recordings/%s/", rfg.Name) + "out_%Y-%m-%d_%H-%M-%S.mp4"

	rfg.recCmd = exec.Command(
		"ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", rfg.URL,
		"-c", "copy",
		"-f", "segment",
		"-segment_time", "300",
		"-reset_timestamps", "1",
		"-strftime", "1",
		dirPath,
	)

	rfg.recCmd.Stdout = os.Stdout
	rfg.recCmd.Stderr = os.Stderr

	if err := rfg.recCmd.Start(); err != nil {
		return ErrFailedToStartCamera
	}

	return nil
}

func (rfg *RtspCapturer) StopRecorder() {
	log.Printf("Stopping recorder for %s", rfg.Name)
	if rfg.recCmd == nil {
		log.Printf("Tried to stop %s but stream does not exist.", rfg.Name)
		return
	}
	rfg.recCmd.Process.Signal(os.Interrupt)
}

func (rfg *RtspCapturer) StartAIStreamer() (io.ReadCloser, error) {
	log.Printf("AI streaming started for: %s\n", rfg.Name)

	rfg.aiCmd = exec.Command(
		"ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", rfg.URL,
		"-vf", fmt.Sprintf("fps=%d,scale=640:480", rfg.Aifps),
		"-c:v", "mjpeg",
		"-f", "image2pipe",
		"pipe:1",
	)

	stdoutPipe, err := rfg.aiCmd.StdoutPipe()
	if err != nil {
		return nil, ErrStdOutError
	}

	rfg.aiCmd.Stderr = os.Stderr
	if err := rfg.aiCmd.Start(); err != nil {
		return nil, ErrFailedToStartCamera
	}

	return stdoutPipe, nil
}

func (rfg *RtspCapturer) StopAIStreamer() {
	rfg.aiCmd.Process.Signal(os.Interrupt)
}
