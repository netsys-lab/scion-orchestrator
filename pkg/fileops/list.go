package fileops

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func ListFilesByPrefixAndSuffix(root string, prefix string, suffix string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, prefix) && strings.HasSuffix(path, suffix) {
			files = append(files, path)
			return nil
		}

		return nil
	})
	return files, err
}
