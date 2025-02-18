package osutils

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
)

// getJournalLogs fetches logs for a systemd service between start and end positions
func GetJournalLogs(service string, lines int) (string, error) {
	// Fetch logs up to 'end' lines
	cmd := exec.Command("journalctl", "-u", service, "--no-hostname", "-o", "cat", "-e", "--lines", strconv.Itoa(lines), "--no-pager")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}
