package scionca

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/netsys-lab/scion-as/pkg/fileops"
	"github.com/netsys-lab/scion-as/pkg/metrics"
)

type CertificateRenewer struct {
	RenewBeforeHours int
	KeyPath          string
	CertPath         string
	ConfigDir        string
	ISDAS            string
	TRC              string
}

func NewCertificateRenewer(configDir string, ISDAS string, renewBeforeHours int) *CertificateRenewer {
	log.Println("[Renewer] Creating new certificate renewer, isdAS ", ISDAS)
	cr := &CertificateRenewer{
		RenewBeforeHours: renewBeforeHours,
		ConfigDir:        configDir,
		ISDAS:            ISDAS,
	}

	return cr
}

func (cr *CertificateRenewer) LoadCertificateFiles() error {
	certDir := filepath.Join(cr.ConfigDir, "crypto", "as")
	cr.KeyPath = filepath.Join(certDir, "cp-as.key")

	certFiles, err := fileops.ListFilesByPrefixAndSuffix(certDir, "ISD", ".pem")
	if err != nil {
		return err
	}

	if len(certFiles) == 0 {
		return fmt.Errorf("No certificate files found in %s", certDir)
	}

	cr.CertPath = certFiles[0]
	isd := strings.Split(cr.ISDAS, "-")[0]

	trcDir := filepath.Join(cr.ConfigDir, "certs")
	trcFiles, err := fileops.ListFilesByPrefixAndSuffix(trcDir, "ISD"+isd, ".trc")
	if err != nil {
		return err
	}
	if len(trcFiles) == 0 {
		return fmt.Errorf("No TRC files found in %s", trcDir)
	}
	sort.Strings(trcFiles)

	cr.TRC = trcFiles[len(trcFiles)-1]
	return nil
}

func (cr *CertificateRenewer) CheckIfCertExpiresSoon() (bool, error) {
	r, _ := ioutil.ReadFile(cr.CertPath)
	block, _ := pem.Decode(r)

	expires := time.Duration(time.Duration(cr.RenewBeforeHours) * (time.Hour))
	deadline := time.Now().Add(expires)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	return deadline.After(cert.NotAfter), nil
}

func (cr *CertificateRenewer) RunRenew() error {

	err := cr.LoadCertificateFiles()
	if err != nil {
		return fmt.Errorf("Failed to load certificate files, %s", err)
	}

	metrics.ASStatus.CertificateRenewal.Status = metrics.SERVICE_STATUS_RUNNING

	log.Println("[Renewer] Checking cert ", cr.CertPath, " to expire within ", cr.RenewBeforeHours, " hours")
	expiresSoon, err := cr.CheckIfCertExpiresSoon()
	if err != nil {
		return fmt.Errorf("[Renewer] Failed to check cert %s for expiration, %s", cr.CertPath, err)
	}

	if !expiresSoon {
		log.Println("[Renewer] Cert is not expiring in the configured deadline, skipping the rest...")
		return nil
	}

	log.Println("[Renewer] Prepare to renew cert ", cr.CertPath, " into tmp dir")
	outCert, err := os.CreateTemp(os.TempDir(), "*.crt")
	if err != nil {
		return err
	}
	outCert.Close()
	os.Remove(outCert.Name())
	outKey, err := os.CreateTemp(os.TempDir(), "*.key")
	if err != nil {
		return err
	}
	outKey.Close()
	os.Remove(outKey.Name())

	log.Println("[Renewer] Renew to cert ", outCert.Name(), " and key ", outKey.Name())
	err = cr.renewCert(outCert.Name(), outKey.Name())
	if err != nil {
		return err
	}

	log.Println("[Renewer] Obtained new cert and key")
	log.Println("[Renewer] Validating new cert")
	err = cr.validateCert(outCert.Name())
	if err != nil {
		return err
	}
	log.Println("[Renewer] Validating done")

	log.Println("[Renewer] Verifying new cert")
	err = cr.verifyCert(outCert.Name())
	if err != nil {
		return err
	}
	log.Println("[Renewer] Verifying done")
	log.Println("[Renewer] Copy tmp files back to original certs")

	err = os.Rename(outCert.Name(), cr.CertPath)
	if err != nil {
		return err
	}

	err = os.Rename(outKey.Name(), cr.KeyPath)
	if err != nil {
		return err
	}
	log.Println("[Renewer] Done")
	return nil
}

func (cr *CertificateRenewer) Run() {
	log.Println("[Renewer] Starting certificate renewal service")
	for {
		err := cr.RunRenew()
		if err != nil {
			log.Println("[Renewer] Failed to renew AS certificate. Error: ", err)
			metrics.ASStatus.CertificateRenewal.Status = metrics.SERVICE_STATUS_ERROR
			metrics.ASStatus.CertificateRenewal.Message = err.Error()
		}
		time.Sleep(1 * time.Hour)
	}
}

func executeCmd(command string, args ...string) (error, string, string) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &out
	log.Printf("[Renewer] Executing: %s\n", cmd.String())
	err := cmd.Run()
	if err == nil {
		log.Printf("[Renewer] Execute successful")
	} else {
		log.Printf("[Renewer] Execute failed %s", err.Error())
	}
	return err, out.String(), stdErr.String()
}

// XXX We might want to include the respective code from scion-pki here later, but for now it's just blocking...
func (cr *CertificateRenewer) validateCert(file string) error {

	err, strOut, strErr := executeCmd("scion-pki", "certificate", "validate", "--type", "chain", file)
	if err != nil {
		return fmt.Errorf("[Renewer] Failed to validate via scion-pki %s, err: %s", err, strErr)
	}
	log.Println("[Renewer] ", strOut)
	return nil
}

func (cr *CertificateRenewer) verifyCert(file string) error {
	err, strOut, strErr := executeCmd("scion-pki", "certificate", "verify", "--trc", cr.TRC, file)
	if err != nil {
		return fmt.Errorf("[Renewer]: Failed to verify via scion-pki %s, err: %s", err, strErr)
	}
	log.Println("[Renewer] ", strOut)
	return nil
}

func (cr *CertificateRenewer) renewCert(outCert string, outKey string) error {
	err, strOut, strErr := executeCmd("scion-pki", "certificate", "renew", cr.CertPath, cr.KeyPath, "--out", outCert, "--out-key", outKey, "--trc", cr.TRC)
	if err != nil {
		return fmt.Errorf("[Renewer]: Failed to renew via scion-pki %s, err: %s", err, strErr)
	}
	log.Println("[Renewer] ", strOut)
	return nil
}
