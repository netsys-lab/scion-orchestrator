package apiv1

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/certutils"
)

const (
	API_PREFIX = "/api/v1/"
)

func RunApiServer(env *environment.HostEnvironment, config *conf.Config) error {
	// Generate a new leaf certificate and private key
	leafCert, leafKey, err := certutils.GenerateLeafCertificate("SCION-Orchestrator API", 365)
	if err != nil {
		log.Fatalf("Failed to generate leaf certificate: %v", err)
	}

	apiFolder := filepath.Join(env.ConfigPath, "api")
	err = os.MkdirAll(apiFolder, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create API folder: %v", err)
	}

	leafCertFile := filepath.Join(apiFolder, "leaf.crt")
	leafKeyFile := filepath.Join(apiFolder, "leaf.key")

	// Save the certificate and key to files
	err = saveCertificate(leafCertFile, leafCert, leafKeyFile, leafKey)
	if err != nil {
		log.Fatalf("Failed to save leaf certificate: %v", err)
	}

	// Start Gin server with the newly generated leaf certificate
	// TODO: locate this file properly
	f, _ := os.Create(filepath.Join(env.LogPath, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	accs := make(gin.Accounts)

	for _, user := range config.Api.Users {
		parts := strings.Split(user, ":")
		accs[parts[0]] = parts[1]
	}

	// Apply the BasicAuth middleware to a specific route group
	authorized := r.Group(API_PREFIX, gin.BasicAuth(accs))

	GenerateCSRFromTemplateHandler(authorized, config.IsdAs, env.ConfigPath)
	AddCertificateChainHandler(authorized, config.IsdAs, env.ConfigPath)
	SignCertificateByCSRHandler(authorized, config.IsdAs, env.ConfigPath, config)

	apiAddress := ":8443"
	if config.Api.Address != "" {
		apiAddress = config.Api.Address
	}

	log.Println("[Api] Starting server on", apiAddress)

	// Start the server with the new certificate and private key
	err = r.RunTLS(apiAddress, leafCertFile, leafKeyFile)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	return nil
}

// saveCertificate saves the certificate and private key to files
func saveCertificate(certPath string, cert *x509.Certificate, keyPath string, key *rsa.PrivateKey) error {
	// Save the certificate to a PEM file
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	err := ioutil.WriteFile(certPath, certPEM, 0644)
	if err != nil {
		return fmt.Errorf("failed to write certificate to file: %v", err)
	}
	log.Println("[Api] Certificate saved to", certPath)

	// Save the private key to a PEM file
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	err = ioutil.WriteFile(keyPath, keyPEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to write private key to file: %v", err)
	}
	log.Println("[Api] Private key saved to", keyPath)

	return nil
}
