package certutils

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/netsys-lab/scion-orchestrator/pkg/fileops"
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

func GetLatestTRCForISD(configDir, isd string) (string, error) {
	trcPath := filepath.Join(configDir, "certs")
	trcFiles, err := fileops.ListFilesByPrefixAndSuffix(trcPath, "ISD"+isd+"-", ".trc")
	sort.Strings(trcFiles)

	if err != nil {
		return "", fmt.Errorf("Failed to list TRC files: %v\n", err)
	}

	if len(trcFiles) == 0 {
		return "", fmt.Errorf("No TRC files found for ISD %s", isd)
	}

	trcFile := trcFiles[len(trcFiles)-1]
	return trcFile, nil
}
