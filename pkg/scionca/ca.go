package scionca

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
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/netsys-lab/scion-as/pkg/fileops"
)

type SCIONCertificateAthority struct {
	ConfigDir     string
	CaCertificate *x509.Certificate
	CaPrivateKey  *ecdsa.PrivateKey
	ISD           string
}

func NewSCIONCertificateAuthority(configDir string, isd string) *SCIONCertificateAthority {

	return &SCIONCertificateAthority{
		ConfigDir: configDir,
		ISD:       isd,
	}
}

func (ca *SCIONCertificateAthority) LoadCA() error {
	// Load CA certificate and private key

	caDir := filepath.Join(ca.ConfigDir, "crypto", "ca")
	keyFile := filepath.Join(caDir, "cp-ca.key")

	caCertFiles, err := fileops.ListFilesByPrefixAndSuffix(caDir, "ISD", ".crt")
	if err != nil {
		return fmt.Errorf("Failed to list CA certificate files: %v\n", err)
	}
	// TODO: We assume only one .crt file here
	certFile := caCertFiles[0]

	caCertPEM, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("Failed to read CA certificate: %v\n", err)

	}
	caKeyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("Failed to read CA private key: %v\n", err)
	}

	caCert, caKey, err := loadCertAndKey(caCertPEM, caKeyPEM)
	if err != nil {
		return fmt.Errorf("Failed to load CA cert and key: %v\n", err)
	}

	ca.CaCertificate = caCert
	ca.CaPrivateKey = caKey
	return nil
}

func (ca *SCIONCertificateAthority) IssueCertificateFromCSR(csrFile string, dstFile, isd string, as string) error {
	log.Println("Issuing certificate from CSR, csrFile ", csrFile)
	csrPEM, err := ioutil.ReadFile(csrFile)
	if err != nil {
		return fmt.Errorf("Failed to read CSR: %v\n", err)
	}
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
					Value: fmt.Sprintf("%s-%s", isd, as),
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
	leafCertDER, err := x509.CreateCertificate(rand.Reader, &leafCertTemplate, ca.CaCertificate, csr.PublicKey, ca.CaPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create leaf certificate: %v", err)
	}

	// Encode the leaf certificate
	leafCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafCertDER})

	// Append the CA certificate to the leaf certificate
	fullChainPEM := append(leafCertPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.CaCertificate.Raw})...)

	// Save the full chain to a file
	err = os.WriteFile(dstFile, fullChainPEM, 0644)
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
