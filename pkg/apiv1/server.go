package apiv1

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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

func SetupCertificates(env *environment.HostEnvironment) (string, string, error) {

	apiFolder := filepath.Join(env.ConfigPath, "api")
	leafCertFile := filepath.Join(apiFolder, "leaf.crt")
	leafKeyFile := filepath.Join(apiFolder, "leaf.key")

	certsNotExistOrExpired := false
	if _, err := os.Stat(leafCertFile); err != nil && os.IsNotExist(err) {
		certsNotExistOrExpired = true
	}

	if !certsNotExistOrExpired && certutils.CheckCertificateExpiration(leafCertFile) != nil {
		certsNotExistOrExpired = true
	}
	log.Println("[Api] Certificates exist and are not expired:", !certsNotExistOrExpired)
	if certsNotExistOrExpired {
		// Generate a new leaf certificate and private key
		leafCert, leafKey, err := certutils.GenerateLeafCertificate("SCION-Orchestrator API", 365)
		if err != nil {
			log.Fatalf("Failed to generate leaf certificate: %v", err)
		}

		err = os.MkdirAll(apiFolder, 0755)
		if err != nil {
			return "", "", fmt.Errorf("failed to create API folder: %v", err)
		}

		// Save the certificate and key to files
		err = saveCertificate(leafCertFile, leafCert, leafKeyFile, leafKey)
		if err != nil {
			log.Fatalf("Failed to save leaf certificate: %v", err)
		}

		log.Println("[Api] New leaf certificate generated and saved")
	}

	return leafCertFile, leafKeyFile, nil

}

func RegisterRoutes(env *environment.HostEnvironment, config *conf.Config, r *gin.Engine) error {

	accs := make(gin.Accounts)

	for _, user := range config.Api.Users {
		parts := strings.Split(user, ":")
		accs[parts[0]] = parts[1]
	}

	// Apply the BasicAuth middleware to a specific route group
	authorized := r.Group(API_PREFIX, gin.BasicAuth(accs))
	authorized.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	GenerateCSRFromTemplateHandler(authorized, config.IsdAs, env.ConfigPath)
	AddCertificateChainHandler(authorized, config.IsdAs, env.ConfigPath)
	SignCertificateByCSRHandler(authorized, config.IsdAs, env.ConfigPath, config)
	AddStatusHandler(authorized)
	AddSettingsHandler(authorized, config)
	GetTopologyHandler(authorized, env.ConfigPath)
	GetSCIONLinksHandler(authorized, env.ConfigPath)
	AddSCIONLinksHandler(authorized, env.ConfigPath)
	GetCertificateChainsHandler(authorized, config.IsdAs, env.ConfigPath)
	GetServiceDetailsHandler(authorized)
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
