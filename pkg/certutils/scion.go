package certutils

import (
	"log"
	"os/exec"
)

// TODO: May import the code from scion-pki in the future
// TODO: Windows Support
func VerifySCIONCertificateChain(certFile string, trcFile string) error {
	cmd := exec.Command("scion-pki", "certificate", "verify", "--trc", trcFile, certFile)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[CertUtils] Could not verify the certificate chain: " + err.Error())
		return err
	}
	return nil
}

// TODO: May import the code from scion-pki in the future
// TODO: Windows Support
func ValidateSCIONCertificateChain(certFile string) error {
	cmd := exec.Command("scion-pki", "certificate", "validate", "--type", "chain", certFile)
	log.Println("[CertUtils] Validating certificate chain, ", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[CertUtils] Could not validate the certificate chain: " + err.Error())
		log.Println("[CertUtils] Output: " + string(out))
		return err
	}
	return nil
}
