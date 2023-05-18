package lib

import (
	"fmt"
	"github.com/kznrluk/aski/config"
	"os"
	"path/filepath"
	"strings"
)

func ReadFileFromPWDAndHistoryDir(partialFilename string) ([]byte, string, error) {
	str := config.MustGetHistoryDir()

	dirsToSearch := []string{".", filepath.Join(str, "")}

	for _, dir := range dirsToSearch {
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, "", err
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), partialFilename) {
				b, err := os.ReadFile(filepath.Join(dir, file.Name()))
				if err != nil {
					return nil, "", fmt.Errorf("cannot read file: %s", err)
				}
				return b, file.Name(), err
			}
		}
	}

	return nil, "", fmt.Errorf("file not found")
}
