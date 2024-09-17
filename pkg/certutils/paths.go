package certutils

import (
	"fmt"
	"path/filepath"
	"strings"
)

func GetASCertificateFilename(configDir, isdAs string) (string, error) {
	parts := strings.Split(isdAs, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid ISD-AS format")
	}

	asnInFile := strings.ReplaceAll(parts[1], ":", "_")
	return filepath.Join(configDir, "crypto", "as", fmt.Sprintf("ISD%s-AS%s.pem", parts[0], asnInFile)), nil
}

func GetASPrivateKeyFilename(configDir string) string {
	return filepath.Join(configDir, "crypto", "as", "cp-as.key")
}
