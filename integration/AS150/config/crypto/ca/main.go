package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

func main() {
	// Load CA certificate and private key
	caCertPEM, err := os.ReadFile("ISD1-AS150.ca.crt")
	if err != nil {
		fmt.Printf("Failed to read CA certificate: %v\n", err)
		return
	}
	caKeyPEM, err := os.ReadFile("cp-ca.key")
	if err != nil {
		fmt.Printf("Failed to read CA private key: %v\n", err)
		return
	}

	caCert, caKey, err := loadCertAndKey(caCertPEM, caKeyPEM)
	if err != nil {
		fmt.Printf("Failed to load CA cert and key: %v\n", err)
		return
	}

	// Load CSR
	csrPEM, err := ioutil.ReadFile("1-50.as.csr")
	if err != nil {
		fmt.Printf("Failed to read CSR: %v\n", err)
		return
	}

	// Issue certificate from CSR
	err = issueCertificateFromCSR(csrPEM, caCert, caKey)
	if err != nil {
		fmt.Printf("Failed to issue certificate from CSR: %v\n", err)
		return
	}

	fmt.Println("Certificate issued from CSR and saved as a full chain.")
}

// issueCertificateFromCSR issues a certificate based on the provided CSR and CA information
func issueCertificateFromCSR(csrPEM []byte, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) error {
	// Parse the CSR
	block, _ := pem.Decode(csrPEM)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return fmt.Errorf("failed to decode CSR PEM")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %v", err)
	}

	// Validate the CSR
	err = csr.CheckSignature()
	if err != nil {
		return fmt.Errorf("CSR signature verification failed: %v", err)
	}

	// Calculate the Subject Key Identifier
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(csr.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	subjectKeyID := sha1.Sum(pubKeyBytes)

	// Create a serial number for the certificate
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	customOID := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 55324, 1, 2, 1}

	// Create the leaf certificate template
	leafCertTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: csr.Subject.CommonName,
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  customOID,
					Value: "1-150",
				},
			},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageTimeStamping},
		SubjectKeyId: subjectKeyID[:],
	}

	// Set the public key from the CSR
	leafCertTemplate.PublicKey = csr.PublicKey

	// Sign the leaf certificate with the CA key
	leafCertDER, err := x509.CreateCertificate(rand.Reader, &leafCertTemplate, caCert, csr.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create leaf certificate: %v", err)
	}

	// Encode the leaf certificate
	leafCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafCertDER})

	// Append the CA certificate to the leaf certificate
	fullChainPEM := append(leafCertPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})...)

	// Save the full chain to a file
	err = os.WriteFile("leaf_cert_full_chain_from_csr.pem", fullChainPEM, 0644)
	if err != nil {
		return fmt.Errorf("failed to write full chain certificate: %v", err)
	}

	return nil
}

// loadCertAndKey loads the certificate and private key from PEM-encoded data
func loadCertAndKey(certPEM, keyPEM []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("failed to decode certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	block, _ = pem.Decode(keyPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("failed to decode private key PEM")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Assert that the parsed key is of type *ecdsa.PrivateKey
	key, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("not an ECDSA private key")
	}

	return cert, key, nil
}
