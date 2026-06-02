package recordingprotocol

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type RtspFFMPEGRecorder struct {
	URL  string
	Name string
	Cmd  *exec.Cmd
}

func CreateNewRtspRecorder(URL string, Name string) *RtspFFMPEGRecorder {
	log.Printf("New recorder for: %s at %s", Name, URL)
	return &RtspFFMPEGRecorder{
		URL,
		Name,
		nil,
	}
}

func (rfg *RtspFFMPEGRecorder) StartStream() error {
	log.Printf("Camera streaming started for: %s\n", rfg.Name)

	if !FolderExist(rfg.Name) {
		log.Printf("Folder not found for %s", rfg.Name)
		return ErrRecordingFolderForCameraNotFound
	}

	dirPath := fmt.Sprintf("recordings/%s/", rfg.Name) + "out_%Y-%m-%d_%H-%M-%S.mp4"

	rfg.Cmd = exec.Command(
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

	rfg.Cmd.Stdout = os.Stdout
	rfg.Cmd.Stderr = os.Stderr

	if err := rfg.Cmd.Run(); err != nil {
		return ErrFailedToStartCamera
	}

	return nil
}

func (rfg *RtspFFMPEGRecorder) StopStream() {
	if rfg.Cmd == nil {
		log.Printf("Tried to stop %s but stream does not exist.", rfg.Name)
		return
	}
	rfg.Cmd.Process.Signal(os.Interrupt)
}
