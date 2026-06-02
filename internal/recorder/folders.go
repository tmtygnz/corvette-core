package recorder

import (
	"fmt"
	"log"
	"os"
)

func FolderExist(name string) bool {
	dirPath := fmt.Sprintf("recordings/%s", name)
	log.Println(dirPath)
	info, err := os.Stat(dirPath)

	if err != nil {
		return false
	}
	return info.IsDir()
}

func SetupCameraFolder(name string) error {
	dirPath := fmt.Sprintf("recordings/%s", name)
	if err := os.Mkdir(dirPath, 0755); err != nil {
		return err
	}
	return nil
}
