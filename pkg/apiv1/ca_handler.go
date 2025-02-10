package apiv1

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/pkg/fileops"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionca"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionutils"
)

func SignCertificateByCSRHandler(eng *gin.RouterGroup, isdAS string, configDir string, config *conf.Config) {

	// TODO: get ISD from isd-as
	// TODO: get certValidity from config
	// TODO: Better use CS API to talk to the CA?
	validityHours := config.Ca.CertValidityHours
	isd := scionutils.GetISDFromISDAS(isdAS)
	ca := scionca.NewSCIONCertificateAuthority(configDir, isd, validityHours)

	eng.POST("ca/certs/:isd/:as/sign", func(c *gin.Context) {
		certIsd := c.Param("isd")
		certAs := c.Param("as")

		if c.Request.Body == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please send a request body"})
			return
		}

		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read the request body"})
			return
		}

		// Get a file name to store the csr
		csrFile, err := fileops.CreateTempFileWithSuffix(".csr")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create the CSR file"})
			return
		}

		_, err = csrFile.Write(bodyBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write the CSR file"})
			return
		}

		err = ca.LoadCA()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the CA"})
			return
		}

		// Get a file name to store the csr
		certFile := fileops.GetTempFileNameWithSuffix(".pem")

		err = ca.IssueCertificateFromCSR(csrFile.Name(), certFile, certIsd, certAs)
		if err != nil {
			log.Println("[CA] Error issuing certificate: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not issue the certificate"})
			return
		}

		c.File(certFile)
	})
}
