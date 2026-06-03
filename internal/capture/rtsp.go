package capture

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
	"os/exec"
)

type CreateNewGenericCapturerOpts struct {
	URL             string
	Name            string
	AiFPS           int
	AiScalingWidth  int
	AiScalingHeight int
}

type RtspCapturer struct {
	URL   string
	Name  string
	Aifps int

	AiFrame   chan []byte
	FrameSize int

	scalingWidth  int
	scalingHeight int

	recCmd *exec.Cmd
	aiCmd  *exec.Cmd
}

func CreateNewCapturer(opts CreateNewGenericCapturerOpts) (*RtspCapturer, error) {
	log.Printf("New recorder for: %s at %s", opts.Name, opts.URL)
	setupFolderForCapturer(opts.Name)

	frameSize := opts.AiScalingWidth * opts.AiScalingHeight * 3

	return &RtspCapturer{
		URL:           opts.URL,
		Name:          opts.Name,
		Aifps:         opts.AiFPS,
		recCmd:        nil,
		aiCmd:         nil,
		FrameSize:     frameSize,
		AiFrame:       make(chan []byte, 1),
		scalingWidth:  opts.AiScalingWidth,
		scalingHeight: opts.AiScalingHeight,
	}, nil
}

func (rfg *RtspCapturer) StartRecorder() error {
	log.Printf("Camera streaming started for: %s\n", rfg.Name)
	dirPath := fmt.Sprintf("recordings/%s/", rfg.Name) + "out_%Y-%m-%d_%H-%M-%S.mp4"

	rfg.recCmd = exec.Command(
		"ffmpeg",
		"-loglevel", "error",

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

func (rfg *RtspCapturer) StartAIStreamer() error {
	log.Printf("AI streaming started for: %s\n", rfg.Name)

	rfg.aiCmd = exec.Command(
		"ffmpeg",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", rfg.URL,
		"-vf", fmt.Sprintf("fps=%d,scale=640:640", rfg.Aifps),
		"-c:v", "rawvideo",
		"-pix_fmt", "rgb24",
		"-f", "rawvideo",
		"pipe:1",
	)

	stdoutPipe, err := rfg.aiCmd.StdoutPipe()
	if err != nil {
		return ErrStdOutError
	}

	rfg.aiCmd.Stderr = os.Stderr
	if err := rfg.aiCmd.Start(); err != nil {
		return ErrFailedToStartCamera
	}

	go rfg.frameToChan(stdoutPipe)

	return nil
}

func (rfg *RtspCapturer) frameToChan(stdout io.ReadCloser) error {
	defer rfg.aiCmd.Wait()

	for {
		frameBuffer := make([]byte, rfg.FrameSize)
		_, err := io.ReadFull(stdout, frameBuffer)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("EOF for stream %s", rfg.Name)
			} else {
				log.Printf("Error reading frame for %s", rfg.Name)
			}
		}
		rfg.AiFrame <- frameBuffer
	}
}

func (rfg *RtspCapturer) StopAIStreamer() {
	rfg.aiCmd.Process.Signal(os.Interrupt)
}

func (rfg *RtspCapturer) GetCurrentAIFrame() ([]byte, bool) {
	frame, ok := <-rfg.AiFrame
	return frame, ok
}

func (rfg *RtspCapturer) getJpeg(data []byte) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, rfg.scalingWidth, rfg.scalingHeight))

	srcIdx := 0
	for y := range rfg.scalingHeight {
		for x := range rfg.scalingWidth {

			r := data[srcIdx]
			g := data[srcIdx+1]
			b := data[srcIdx+2]
			srcIdx += 3

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var imageBuf bytes.Buffer
	err := jpeg.Encode(&imageBuf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Println("Failed to generate jpeg")
		return nil, err
	}

	return imageBuf.Bytes(), nil
}
