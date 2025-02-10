package fileops

import (
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandStringBytesMaskImprSrc returns a random hexadecimal string of length n.
func randomString(n int) string {
	b := make([]byte, (n+1)/2) // can be simplified to n/2 if n is always even

	if _, err := src.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[:n]
}

func GetTempFileNameWithSuffix(suffix string) string {
	dir := os.TempDir()
	return filepath.Join(dir, randomString(16)+suffix)
}

func CreateTempFileWithSuffix(suffix string) (*os.File, error) {
	dir := os.TempDir()
	return os.Create(filepath.Join(dir, randomString(16)+suffix))
}
