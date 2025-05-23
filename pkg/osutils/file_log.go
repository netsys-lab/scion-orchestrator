package osutils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// GetFileLogs efficiently reads the last 'lines' number of log entries from a file without loading the full file into memory
func GetFileLogs(filePath string, lines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file stats: %w", err)
	}

	// Read from the end of the file
	var offset int64 = stat.Size()
	var count int
	var result []string

	// Read in reverse
	buf := make([]byte, 1) // Read one byte at a time
	for offset > 0 {
		offset--
		_, err := file.Seek(offset, io.SeekStart) // Move backwards
		if err != nil {
			return "", fmt.Errorf("error seeking file: %w", err)
		}

		_, err = file.Read(buf)
		if err != nil {
			return "", fmt.Errorf("error reading file: %w", err)
		}

		// If we find a newline, process the line
		if buf[0] == '\n' {
			// Read the line
			line, err := readLine(file)
			if err == nil {
				result = append([]string{line}, result...) // Prepend to the slice
				count++
			}

			// Stop if we have enough lines
			if count >= lines {
				break
			}
		}
	}

	// Read first line if the file has less than `lines` lines
	if count < lines {
		file.Seek(0, io.SeekStart)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			result = append(result, scanner.Text())
		}
	}

	if len(result) == 0 {
		return "No logs available.", nil
	}

	return joinLines(result), nil
}

// readLine reads a single line from a file at the current position
func readLine(file *os.File) (string, error) {
	var line string
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line = scanner.Text()
	}
	return line, scanner.Err()
}

// joinLines is a helper function to join log lines into a single string
func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
