package apiv1

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/pkg/certutils"
	"github.com/netsys-lab/scion-orchestrator/pkg/fileops"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionutils"
)

func GetCertificateChainsHandler(eng *gin.RouterGroup, isdAS string, configDir string) {

	eng.GET("cppki/certs", func(c *gin.Context) {
		certs, err := certutils.GetASCertificateDetails(configDir, isdAS)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get the certificate chain"})
			return
		}

		c.JSON(http.StatusOK, certs)
	})

}

func GenerateCSRFromTemplateHandler(eng *gin.RouterGroup, isdAS string, configDir string) {

	eng.POST("cppki/csr", func(c *gin.Context) {

		var payload ApiCSR

		// Bind the incoming JSON to the RequestPayload struct
		if err := c.BindJSON(&payload); err != nil {
			// If binding fails, return an error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate a csr.tmpl file with the json content
		// Serialise json into a file
		// Create or open the file for writing
		csrTemplateFile, err := os.CreateTemp("/tmp/", "*.csr.tmpl")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create the CSR template"})
			return
		}
		err = json.NewEncoder(csrTemplateFile).Encode(payload.Subject)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store the CSR template"})
		}
		csrTemplateFile.Close()

		// Create csr file, but close and remove it immediatly since it will be written by scion-pki
		// TODO: Create fileops.RandomFile() function
		csrFile, err := os.CreateTemp("/tmp/", "*.csr")
		csrFile.Close()
		os.Remove(csrFile.Name())

		privateKey := certutils.GetASPrivateKeyFilename(configDir)

		cmd := exec.Command("scion-pki", "certificate", "create", "--csr", "--key", privateKey, csrTemplateFile.Name(), csrFile.Name())
		_, err = cmd.CombinedOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate the CSR"})
			return
		}

		c.File(csrFile.Name())
	})

}

func AddCertificateChainHandler(eng *gin.RouterGroup, isdAS string, configDir string) {

	eng.POST("cppki/certs", func(c *gin.Context) {
		// Require content type application/x-pem-files
		/*
			 * Input certificate chain
			 -----BEGIN CERTIFICATE-----
				ASCertificate ...
			 -----END CERTIFICATE-----
			 -----BEGIN CERTIFICATE-----
				CACertificate ...
			 -----END CERTIFICATE-----
		*/

		log.Println("[CPPKI] Adding certificate chain")

		// Read the body
		// Write the body to a file
		if c.Request.Body == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please send a request body"})
			return
		}

		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read the request body"})
			return
		}

		certChainFile, err := fileops.CreateTempFileWithSuffix(".pem")
		// os.CreateTemp("/tmp/", "*.pem")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create the certificate chain file"})
			return
		}

		bodyBytes = formatPEMString(string(bodyBytes))

		_, err = certChainFile.Write(bodyBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write the certificate chain file"})
			return
		}

		err = certutils.ValidateSCIONCertificateChain(certChainFile.Name())
		if err != nil {
			log.Println("[CPPKI] Could not validate the certificate chain: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not validate the certificate chain"})
			return
		}

		isd := scionutils.GetISDFromISDAS(isdAS)

		trcFile, err := certutils.GetLatestTRCForISD(configDir, isd)
		if err != nil {
			log.Println("[CPPKI] Could not get the latest TRC file: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get the latest TRC file"})
			return
		}

		err = certutils.VerifySCIONCertificateChain(certChainFile.Name(), trcFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify the certificate chain"})
			return
		}

		// Move the certificate chain to the correct location
		certFileName, err := certutils.GetASCertificateFilename(configDir, isdAS)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get the AS certificate filename"})
			return
		}

		if fileops.FileOrFolderExists(certFileName) {
			err = os.Rename(certFileName, certFileName+".old")
			if err != nil {
				log.Println("[CPPKI] Could not backup the old AS certificate: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not backup the old AS certificate"})
				return
			}
		}

		err = os.Rename(certChainFile.Name(), certFileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not overwrite the AS certificate"})
			return
		}

		log.Println("[CPPKI] Certificate chain added successfully")

		c.JSON(http.StatusOK, gin.H{"message": "Certificate chain added successfully"})

	})
}

func formatPEMString(pemStr string) []byte {
	// Remove all existing newlines to process as a single string
	pemStr = strings.ReplaceAll(pemStr, "\n", "")

	// Regular expression to match multiple PEM certificates
	pemRegex := regexp.MustCompile(`(-----BEGIN CERTIFICATE-----)([A-Za-z0-9+/=]+)(-----END CERTIFICATE-----)`)

	// Buffer to store formatted PEM output
	var formattedPem bytes.Buffer

	// Find all certificate blocks
	matches := pemRegex.FindAllStringSubmatch(pemStr, -1)
	if len(matches) == 0 {
		return nil // Return nil if no valid PEM found
	}

	for _, match := range matches {
		beginMarker := match[1]
		pemData := match[2]
		endMarker := match[3]

		// Reformat certificate block by inserting newlines every 64 characters
		formattedPem.WriteString(beginMarker + "\n")
		for i := 0; i < len(pemData); i += 64 {
			end := i + 64
			if end > len(pemData) {
				end = len(pemData)
			}
			formattedPem.WriteString(pemData[i:end] + "\n")
		}
		formattedPem.WriteString(endMarker + "\n") // Separate multiple certs
	}

	return formattedPem.Bytes()
}
