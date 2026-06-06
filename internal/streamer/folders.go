package streamer

import (
	"fmt"
	"log/slog"
	"os"
)

func getPath(name string) string {
	dirPath := fmt.Sprintf("./recordings/%s", name)
	return dirPath
}

func DoesThisFolderExist(name string) bool {
	_, err := os.Stat(getPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func CreateFolder(name string) error {
	err := os.Mkdir(getPath(name), 0750)
	if err != nil && !os.IsExist(err) {
		slog.Error("Failed to create folder.", "for", name)
		return err
	}

	return nil
}
