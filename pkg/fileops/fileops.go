package fileops

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(dst string, src string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Flush the write buffer to ensure all data is written to the destination file
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush write buffer: %w", err)
	}

	return nil
}

func CopyDir(dst string, src fs.FS, root string) error {
	return fs.WalkDir(src, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}

		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, os.ModePerm)
	})
}

func ReplaceStringInFile(filePath, oldString, newString string) error {
	// Read the file into memory
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Replace the string
	newContent := strings.ReplaceAll(string(fileContent), oldString, newString)

	// Write the modified content back to the file
	return ioutil.WriteFile(filePath, []byte(newContent), 0644)
}

func ReplaceSingleBackslashWithDouble(filePath string) error {
	// Read the file into memory
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Replace single backslash with double backslashes
	newContent := strings.ReplaceAll(string(fileContent), "\\", "\\\\")

	// Write the modified content back to the file
	return ioutil.WriteFile(filePath, []byte(newContent), 0644)
}
