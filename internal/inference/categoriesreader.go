package inference

import (
	"os"
	"strings"
)

func ReadCategories(filePath string) []string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	dataStr := string(data)
	arr := strings.Split(dataStr, "\n")
	return arr
}
