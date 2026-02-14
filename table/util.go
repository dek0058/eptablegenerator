package table

import (
	"os"
	"path/filepath"
)

func FindXLSX(dirPath string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".xlsx" {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}
