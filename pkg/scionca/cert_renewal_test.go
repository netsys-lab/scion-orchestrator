package scionca

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func createTempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "scionca_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tempDir
}

func setupTestDirectories(t *testing.T, tempDir string) (string, string) {
	certDir := filepath.Join(tempDir, "crypto", "as")
	trcDir := filepath.Join(tempDir, "certs")

	if err := os.MkdirAll(certDir, 0755); err != nil {
		t.Fatalf("Failed to create cert dir: %v", err)
	}
	if err := os.MkdirAll(trcDir, 0755); err != nil {
		t.Fatalf("Failed to create trc dir: %v", err)
	}

	return certDir, trcDir
}

func createTRCFile(t *testing.T, trcDir, isd string, base, serial int) string {
	trcPath := filepath.Join(trcDir, fmt.Sprintf("ISD%s-B%d-S%d.trc", isd, base, serial))
	trcData := []byte("dummy TRC data")
	if err := ioutil.WriteFile(trcPath, trcData, 0644); err != nil {
		t.Fatalf("Failed to write TRC file: %v", err)
	}
	return trcPath
}

func TestNewCertificateRenewer(t *testing.T) {
	cr := NewCertificateRenewer("/config", "1-ff00:0:110", 24)
	if cr.ConfigDir != "/config" {
		t.Errorf("Expected ConfigDir to be /config, got %s", cr.ConfigDir)
	}
	if cr.ISDAS != "1-ff00:0:110" {
		t.Errorf("Expected ISDAS to be 1-ff00:0:110, got %s", cr.ISDAS)
	}
	if cr.RenewBeforeHours != 24 {
		t.Errorf("Expected RenewBeforeHours to be 24, got %d", cr.RenewBeforeHours)
	}
	if cr.listFilesByPrefixAndSuffix == nil {
		t.Error("Expected listFilesByPrefixAndSuffix to be set")
	}
}
func TestLoadCertificateFiles(t *testing.T) {
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	certDir, trcDir := setupTestDirectories(t, tempDir)
	isd := "1"
	isdas := isd + "-ff00:0:110"

	// Create test files
	certPath := filepath.Join(certDir, "ISD"+isd+"-B1-S1.pem")
	trcPath1 := createTRCFile(t, trcDir, isd, 1, 1)
	trcPath9 := createTRCFile(t, trcDir, isd, 1, 9)
	trcPath10 := createTRCFile(t, trcDir, isd, 1, 10)

	tests := []struct {
		name          string
		trcFiles      []string
		expectedTRC   string
		expectSuccess bool
	}{
		{
			name:          "Only trcPath1",
			trcFiles:      []string{trcPath1},
			expectedTRC:   trcPath1,
			expectSuccess: true,
		},
		{
			name:          "Both trcPath9 and trcPath10",
			trcFiles:      []string{trcPath10, trcPath9},
			expectedTRC:   trcPath10,
			expectSuccess: true,
		},
		{
			name:          "No TRC files",
			trcFiles:      []string{},
			expectedTRC:   "",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the file listing function
			mockList := func(root, prefix, suffix string) ([]string, error) {
				if root == certDir && prefix == "ISD" && suffix == ".pem" {
					return []string{certPath}, nil
				}
				if root == trcDir && prefix == "ISD"+isd && suffix == ".trc" {
					return tt.trcFiles, nil
				}
				return nil, fmt.Errorf("unexpected arguments")
			}

			cr := &CertificateRenewer{
				ConfigDir:                  tempDir,
				ISDAS:                      isdas,
				listFilesByPrefixAndSuffix: mockList,
			}

			err := cr.LoadCertificateFiles()
			if tt.expectSuccess {
				if err != nil {
					t.Fatalf("LoadCertificateFiles failed: %v", err)
				}

				expectedKeyPath := filepath.Join(certDir, "cp-as.key")
				if cr.KeyPath != expectedKeyPath {
					t.Errorf("Expected KeyPath to be %s, got %s", expectedKeyPath, cr.KeyPath)
				}
				if cr.CertPath != certPath {
					t.Errorf("Expected CertPath to be %s, got %s", certPath, cr.CertPath)
				}
				if cr.TRC != tt.expectedTRC {
					t.Errorf("Expected TRC to be %s, got %s", tt.expectedTRC, cr.TRC)
				}
			} else if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}
