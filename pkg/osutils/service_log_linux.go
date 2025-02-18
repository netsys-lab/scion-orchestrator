package osutils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// getJournalLogs fetches logs for a systemd service between start and end positions
func GetJournalLogs(service string, start, end int) (string, error) {
	// Ensure valid range
	if start >= end {
		return "", fmt.Errorf("start (%d) must be less than end (%d)", start, end)
	}

	// Fetch logs up to 'end' lines
	cmd := exec.Command("journalctl", "-u", service, "--lines", strconv.Itoa(end), "--no-pager")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Extract only the required range
	lines := strings.Split(out.String(), "\n")
	if start >= len(lines) {
		return "", fmt.Errorf("start index out of range")
	}

	// Slice to return only the desired log range
	selectedLogs := lines[start:end]
	return strings.Join(selectedLogs, "\n"), nil
}
