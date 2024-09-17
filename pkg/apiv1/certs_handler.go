package apiv1

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/pkg/certutils"
)

func GenerateCSRFromTemplateHandler(eng *gin.RouterGroup, isdAS string, configDir string) {

	eng.POST("csr", func(c *gin.Context) {

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

		// Create csr file, but close it immediatly since it will be written by scion-pki
		csrFile, err := os.CreateTemp("/tmp/", "*.csr")
		csrFile.Close()

		os.Remove(csrFile.Name())

		privateKey := certutils.GetASPrivateKeyFilename(configDir)

		cmd := exec.Command("scion-pki", "certificate", "create", "--csr", "--key", privateKey, csrTemplateFile.Name(), csrFile.Name())
		//log.Println(cmd.String())
		_, err = cmd.CombinedOutput()
		//log.Println(string(result))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate the CSR"})
			return
		}

		c.File(csrFile.Name())

		// Run scion-pki
		// scion-pki certificate create --csr  --key step-ca/71-88.as.key step-ca/71-88.csr.tmpl step-ca/71-88.as.csr

		// c.JSON(http.StatusOK, gin.H{"csr": csr})
	})

}
