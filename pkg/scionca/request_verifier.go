package scionca

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"time"

	"github.com/scionproto/scion/pkg/addr"
	"github.com/scionproto/scion/pkg/private/serrors"
	"github.com/scionproto/scion/pkg/scrypto/cms/protocol"
	"github.com/scionproto/scion/pkg/scrypto/cppki"
)

func decodePem(certInput []byte) tls.Certificate {
	var cert tls.Certificate
	certPEMBlock := []byte(certInput)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return cert
}

func VerifyCMSSignedRenewalRequest(ctx context.Context,
	req []byte, r *RequestVerifier) (*x509.CertificateRequest, error) {

	ci, err := protocol.ParseContentInfo(req)
	if err != nil {
		return nil, serrors.Wrap("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.Wrap("parsing SignedData", err)
	}

	chain, err := ExtractChain(sd)
	if err != nil {
		return nil, serrors.Wrap("extracting signing certificate chain", err)
	}

	if err := r.VerifySignature(ctx, sd, chain); err != nil {
		return nil, err
	}

	pld, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return nil, serrors.Wrap("reading payload", err)
	}

	csr, err := x509.ParseCertificateRequest(pld)
	if err != nil {
		return nil, serrors.Wrap("parsing CSR", err)
	}

	return csr, nil // r.processCSR(csr, chain[0])
}

func ExtractAndVerifyCsr(trcPath string, bts []byte, file *os.File) (*x509.CertificateRequest, error) {
	r := RequestVerifier{
		TRCFetcher: &LocalFetcher{
			TrcPath: trcPath,
		},
	}

	csr, err := VerifyCMSSignedRenewalRequest(context.Background(), bts, &r)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		// return nil, err
	}

	err = pem.Encode(file, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr.Raw})
	if err != nil {
		return nil, err
	}
	return csr, nil
}

// DecodeSignedTRC parses the signed TRC.
func DecodeSignedTRC(raw []byte) (cppki.SignedTRC, error) {
	ci, err := protocol.ParseContentInfo(raw)
	if err != nil {
		return cppki.SignedTRC{}, serrors.Wrap("error parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return cppki.SignedTRC{}, serrors.Wrap("error parsing SignedData", err)
	}
	if sd.Version != 1 {
		return cppki.SignedTRC{}, serrors.New("unsupported SignedData version", "version", 1)
	}
	if !sd.EncapContentInfo.IsTypeData() {
		return cppki.SignedTRC{}, serrors.Wrap("unsupported EncapContentInfo type", err,
			"type", sd.EncapContentInfo.EContentType)
	}
	praw, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return cppki.SignedTRC{}, serrors.Wrap("error reading raw payload", err)
	}
	trc, err := cppki.DecodeTRC(praw)
	if err != nil {
		return cppki.SignedTRC{}, serrors.Wrap("error parsing TRC payload", err)
	}
	return cppki.SignedTRC{Raw: raw, TRC: trc, SignerInfos: sd.SignerInfos}, nil
}

type LocalFetcher struct {
	TrcPath string
}

func NewLocalFetcher(trcPath string) *LocalFetcher {
	return &LocalFetcher{
		TrcPath: trcPath,
	}

}

func (lf *LocalFetcher) SignedTRC(ctx context.Context, isd addr.ISD) (cppki.SignedTRC, error) {

	trc := cppki.SignedTRC{}

	// Load local trcs and find the one with the highest number
	/*files, err := ioutil.ReadDir(lf.TrcPath)
	if err != nil {
		return trc, fmt.Errorf("Can not read trc files in %s: %v", lf.TrcPath, err)
	}

	if len(files) == 0 {
		return trc, fmt.Errorf("No trc files in %s", lf.TrcPath)
	}

	fileNames := make([]string, 0)
	for _, f := range files {
		baseName := filepath.Base(f.Name())
		if strings.Index(baseName, isd.String()) >= 0 {
			fileNames = append(fileNames, baseName)
		}
	}

	sort.Sort(sort.StringSlice(fileNames))
	trcId := fileNames[len(fileNames)-1]*/

	// trcFile := filepath.Join(lf.TrcPath, trcId)
	trcFile := lf.TrcPath
	log.Println("[CA] Reading TRC ", trcFile)
	bts, err := os.ReadFile(trcFile)
	if err != nil {
		return trc, nil
	}

	block, _ := pem.Decode(bts)
	sTrc, err := DecodeSignedTRC(block.Bytes)
	if err != nil {
		return trc, err
	}
	return sTrc, nil
}

type TRCFetcher interface {
	// SignedTRC fetches the signed TRC for a given ID.
	// The latest TRC can be requested by setting the serial and base number
	// to scrypto.LatestVer.
	SignedTRC(ctx context.Context, isd addr.ISD) (cppki.SignedTRC, error)
}

type RequestVerifier struct {
	TRCFetcher TRCFetcher
}

// VerifyCMSSignedRenewalRequest verifies a renewal request that is encapsulated in a CMS
// envelop. It checks that the contained CSR is valid and correctly self-signed, and
// that the signature is valid and can be verified by the chain included in the CMS envelop.
func (r RequestVerifier) VerifyCMSSignedRenewalRequest(ctx context.Context,
	req []byte) (*x509.CertificateRequest, error) {

	ci, err := protocol.ParseContentInfo(req)
	if err != nil {
		return nil, serrors.Wrap("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.Wrap("parsing SignedData", err)
	}

	chain, err := ExtractChain(sd)
	if err != nil {
		return nil, serrors.Wrap("extracting signing certificate chain", err)
	}

	if err := r.VerifySignature(ctx, sd, chain); err != nil {
		return nil, err
	}

	pld, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return nil, serrors.Wrap("reading payload", err)
	}

	csr, err := x509.ParseCertificateRequest(pld)
	if err != nil {
		return nil, serrors.Wrap("parsing CSR", err)
	}

	return r.processCSR(csr, chain[0])
}

// VerifySignature verifies the signature on the signed data with the provided
// chain. It is checked that the certificate chain is verifiable with an
// active TRC, and that the signature can be verified with the chain.
func (r RequestVerifier) VerifySignature(
	ctx context.Context,
	sd *protocol.SignedData,
	chain []*x509.Certificate,
) error {

	if sd.Version != 1 {
		return serrors.New("unsupported SignedData version", "actual", sd.Version, "supported", 1)
	}
	if c := len(sd.SignerInfos); c != 1 {
		return serrors.New("unexpected number of SignerInfos", "count", c)
	}
	si := sd.SignerInfos[0]
	signer, err := si.FindCertificate(chain)
	if err != nil {
		return serrors.Wrap("selecting client certificate", err)
	}
	if signer != chain[0] {
		return serrors.New("not signed with AS certificate",
			"common_name", signer.Subject.CommonName)
	}
	if err := r.verifyClientChain(ctx, chain); err != nil {
		return serrors.Wrap("verifying client chain", err)
	}

	if !sd.EncapContentInfo.IsTypeData() {
		return serrors.New("unsupported EncapContentInfo type",
			"type", sd.EncapContentInfo.EContentType)
	}
	pld, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return serrors.Wrap("reading payload", err)
	}

	if err := verifySignerInfo(pld, chain[0], si); err != nil {
		return serrors.Wrap("verifying signer info", err)
	}

	return nil
}

func (r RequestVerifier) verifyClientChain(ctx context.Context, chain []*x509.Certificate) error {
	ia, err := cppki.ExtractIA(chain[0].Subject)
	if err != nil {
		return err
	}

	trc, err := r.TRCFetcher.SignedTRC(ctx, ia.ISD())
	if err != nil {
		return serrors.Wrap("loading TRC to verify client chain", err)
	}
	if trc.IsZero() {
		return serrors.New("TRC not found", "isd", ia.ISD())
	}
	now := time.Now()
	if val := trc.TRC.Validity; !val.Contains(now) {
		return serrors.New("latest TRC currently not active", "validity", val, "current_time", now)
	}
	opts := cppki.VerifyOptions{TRC: []*cppki.TRC{&trc.TRC}}
	if err := cppki.VerifyChain(chain, opts); err != nil {
		// If the the previous TRC is in grace period the CA certificate of the chain might
		// have been issued with a previous Root. Try verifying with the TRC in grace period.
		if now.After(trc.TRC.GracePeriodEnd()) {
			return serrors.Wrap("verifying client chain", err)
		}
		graceID := trc.TRC.ID
		graceID.Serial--
		if err := r.verifyWithGraceTRC(ctx, now, graceID.ISD, chain); err != nil {
			return serrors.Wrap("verifying client chain with TRC in grace period "+
				"after verification failure with latest TRC", err,
				"trc_id", trc.TRC.ID,
				"grace_trc_id", graceID,
			)
		}

	}
	return nil
}

func (r RequestVerifier) verifyWithGraceTRC(
	ctx context.Context,
	now time.Time,
	id addr.ISD,
	chain []*x509.Certificate,
) error {

	trc, err := r.TRCFetcher.SignedTRC(ctx, id)
	if err != nil {
		return serrors.Wrap("loading TRC in grace period", err)
	}
	if trc.IsZero() {
		return serrors.New("TRC in grace period not found")
	}
	if val := trc.TRC.Validity; !val.Contains(now) {
		return serrors.New("TRC in grace period not active",
			"validity", val,
			"current_time", now,
		)
	}
	verifyOptions := cppki.VerifyOptions{TRC: []*cppki.TRC{&trc.TRC}}
	if err := cppki.VerifyChain(chain, verifyOptions); err != nil {
		return serrors.Wrap("verifying client chain", err)
	}
	return nil
}

func verifySignerInfo(pld []byte, cert *x509.Certificate, si protocol.SignerInfo) error {
	hash, err := si.Hash()
	if err != nil {
		return err
	}
	attrDigest, err := si.GetMessageDigestAttribute()
	if err != nil {
		return err
	}
	actualDigest := hash.New()
	actualDigest.Write(pld)
	if !bytes.Equal(attrDigest, actualDigest.Sum(nil)) {
		return serrors.New("message digest does not match")
	}
	sigInput, err := si.SignedAttrs.MarshaledForVerifying()
	if err != nil {
		return err
	}
	algo := si.X509SignatureAlgorithm()
	return cert.CheckSignature(algo, sigInput, si.Signature)
}

func (r RequestVerifier) processCSR(csr *x509.CertificateRequest,
	cert *x509.Certificate) (*x509.CertificateRequest, error) {

	csrIA, err := cppki.ExtractIA(csr.Subject)
	if err != nil {
		return nil, serrors.Wrap("extracting ISD-AS from CSR", err)
	}
	chainIA, err := cppki.ExtractIA(cert.Subject)
	if err != nil {
		return nil, serrors.Wrap("extracting ISD-AS from certificate chain", err)
	}
	if !csrIA.Equal(chainIA) {
		return nil, serrors.New("signing subject is different from CSR subject",
			"csr_isd_as", csrIA, "chain_isd_as", chainIA)
	}
	if err := csr.CheckSignature(); err != nil {
		return nil, serrors.Wrap("invalid CSR signature", err)
	}
	return csr, nil
}

func ExtractChain(sd *protocol.SignedData) ([]*x509.Certificate, error) {
	certs, err := sd.X509Certificates()
	if err == nil {
		if len(certs) == 0 {
			err = protocol.ErrNoCertificate
		} else if len(certs) != 2 {
			err = serrors.New("unexpected number of certificates", "count", len(certs))
		}
	}
	if err != nil {
		return nil, serrors.Wrap("parsing certificate chain", err)
	}

	certType, err := cppki.ValidateCert(certs[0])
	if err != nil {
		return nil, serrors.Wrap("checking certificate type", err)
	}
	if certType == cppki.CA {
		certs[0], certs[1] = certs[1], certs[0]
	}
	if err := cppki.ValidateChain(certs); err != nil {
		return nil, serrors.Wrap("validating chain", err)
	}
	return certs, nil
}
