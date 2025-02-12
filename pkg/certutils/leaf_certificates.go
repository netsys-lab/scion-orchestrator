package certutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// generateLeafCertificate generates a self-signed leaf certificate with a new private key
func GenerateLeafCertificate(commonName string, validityDays int) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate a new RSA private key
	privateKey, err := GenerateRsaPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create a serial number for the certificate
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	// Create a certificate template for the leaf certificate
	certTemplate := x509.Certificate{
		SerialNumber: serialNumber, // Unique serial number
		Subject: pkix.Name{ // Subject details
			CommonName:   commonName,                    // Common Name for the certificate (e.g., localhost)
			Organization: []string{"Your Organization"}, // Organization name
		},
		NotBefore:             time.Now(),                                                   // Valid from the current time
		NotAfter:              time.Now().Add(time.Duration(validityDays) * 24 * time.Hour), // Certificate's validity period
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, // Ext key usage (e.g., for server authentication)
		IsCA:                  false,                                          // This is a leaf certificate, not a CA
		BasicConstraintsValid: true,                                           // Required for a valid certificate
	}

	// Create a self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Parse the DER-encoded certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, privateKey, nil
}

// saveCertificate saves the certificate and private key to files
func SaveCertificate(certPath string, cert *x509.Certificate, keyPath string, key *rsa.PrivateKey) error {
	// Save the certificate to a PEM file
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	err := os.WriteFile(certPath, certPEM, 0644)
	if err != nil {
		return fmt.Errorf("failed to write certificate to file: %v", err)
	}
	fmt.Println("Certificate saved to", certPath)

	// Save the private key to a PEM file
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	err = os.WriteFile(keyPath, keyPEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to write private key to file: %v", err)
	}
	fmt.Println("Private key saved to", keyPath)

	return nil
}

// CheckCertificateExpiration reads a PEM-encoded certificate and checks if it's expired
func CheckCertificateExpiration(certPath string) error {
	// Read the certificate file
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	// Decode the PEM block
	block, _ := pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("failed to parse PEM block from file")
	}

	// Parse the X.509 certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Get current time
	now := time.Now()

	// Check expiration
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate expired on %v", cert.NotAfter)
	}

	// fmt.Printf("Certificate is valid. Expiration date: %v\n", cert.NotAfter)
	return nil
}
