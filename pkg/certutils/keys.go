package certutils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GeneratePrivateKey generates a new private key.
func GenerateEcdsaPrivateKey(curve string) (*ecdsa.PrivateKey, error) {
	switch strings.ToLower(curve) {
	case "p-256", "p256":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "p-384", "p384":
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "p-521", "p521":
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("unsupported curve for private ecdsa key %s", curve)
	}
}

func GenerateRsaPrivateKey() (*rsa.PrivateKey, error) {
	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	return privateKey, nil
}

// EncodePEMPrivateKey encodes the private key in PEM format.
func EncodeEcdsaPEMPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	raw, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: raw,
	})
	if p == nil {
		return nil, fmt.Errorf("PEM encoding failed")
	}
	return p, nil
}

func EnsureASPrivateKeyExists(configDir, isdAs string) error {

	keyPath := filepath.Join(configDir, "crypto", "as", "cp-as.key")
	log.Println("[Crypto] Checking if AS private key exists at", keyPath)

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Println("[Crypto] AS private key does not exist. Generating a new ECDSA key with p-256 ...")
		// Generate a new AS private key
		/*key, err := GenerateEcdsaPrivateKey("p256")
		if err != nil {
			return fmt.Errorf("failed to generate AS private key: %v", err)
		}

		// Encode the private key in PEM format
		keyPEM, err := EncodeEcdsaPEMPrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to encode AS private key: %v", err)
		}

		// Save the private key to a file
		err = WritePrivateKeyToFile(keyPEM, keyPath)
		if err != nil {
			return fmt.Errorf("failed to save AS private key: %v", err)
		}*/

		cmd := exec.Command("scion-pki", "key", "private", keyPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create AS private key: %v: %s", err, string(out))
		}

		log.Println("[Crypto] New AS private key saved to", keyPath)
	} else {
		log.Println("[Crypto] AS private key already exists, checking if it is valid ...")
		// Load the existing AS private key
		_, err := LoadPrivateKey(keyPath)
		if err != nil {
			return fmt.Errorf("failed to load AS private key: %v", err)
		}
		log.Println("[Crypto] AS private key is valid")
	}

	return nil

}

// LoadPrivate key loads a private key from file.
func LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading private key %s", err)
	}
	p, rest := pem.Decode(raw)
	if p == nil {
		return nil, fmt.Errorf("parsing private key failed")
	}
	if len(rest) != 0 {
		return nil, fmt.Errorf("file must only contain private key")
	}
	if p.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("file does not contain a private key type but %s", p.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key %s", err)
	}

	priv, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("cannot cast to ECDSA private key %s", fmt.Sprintf("%T", key))
	}
	return priv, nil
}

func WritePrivateKeyToFile(pemEncodedData []byte, file string) error {

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return fmt.Errorf("creating directory for private key: %w", err)
	}

	if err := os.WriteFile(file, pemEncodedData, 0600); err != nil {
		return fmt.Errorf("writing private key to file: %w", err)
	}
	return nil
}
