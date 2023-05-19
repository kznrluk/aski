package file

import (
	"github.com/kznrluk/aski/util"
	"os"
	"path/filepath"
)

type FileContents struct {
	Name     string
	Path     string
	Contents string
	Length   int
}

func GetFileContents(fileGlobs []string) []FileContents {
	var fileContents []FileContents
	for _, arg := range fileGlobs {
		files, err := filepath.Glob(arg)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			contentsBytes, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			content := string(contentsBytes)
			if util.IsBinary(contentsBytes) {
				continue
			}

			info, err := os.Stat(file)
			if err != nil {
				panic(err)
			}

			fileContents = append(fileContents, FileContents{
				Name:     info.Name(),
				Path:     file,
				Contents: content,
				Length:   len(content),
			})
		}
	}
	return fileContents
}
