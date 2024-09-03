package scionca

// This file is auto-generated, don't modify it manually

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/pkg/fileops"
	"github.com/netsys-lab/scion-as/pkg/scionca/models"

	caconfig "github.com/scionproto/scion/private/ca/config"
)

// NewRouter creates a new router for the spec and the given handlers.
// CA Service
//
// API for renewing SCION certificates.
//
// 0.1.0
//

type CaClient struct {
	ClientId string
	Secret   *caconfig.PEMSymmetricKey
}

type CaApiServer struct {
	LatestTRC string
	CaConfig  *conf.CA
	Router    http.Handler
	Clients   []CaClient
	CA        *SCIONCertificateAthority
}

func NewCaApiServer(configDir string, caConfig *conf.CA, ca *SCIONCertificateAthority) *CaApiServer {
	r := chi.NewRouter()

	ar := &CaApiServer{
		Router:   r,
		Clients:  []CaClient{},
		CA:       ca,
		CaConfig: caConfig,
	}

	r.Get("/healthcheck", healthCheck)
	// r.Post("/auth/token", ar.auth)
	r.Post("/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", ar.renewCert)

	return ar
}

func (ar *CaApiServer) loadLatestTRC() error {
	trcPath := filepath.Join(ar.CA.ConfigDir, "certs")
	trcFiles, err := fileops.ListFilesByPrefixAndSuffix(trcPath, "ISD"+ar.CA.ISD+"-", ".trc")
	sort.Strings(trcFiles)

	if err != nil {
		return fmt.Errorf("Failed to list TRC files: %v\n", err)
	}

	if len(trcFiles) == 0 {
		return fmt.Errorf("No TRC files found for ISD %s", ar.CA.ISD)
	}

	trcFile := trcFiles[len(trcFiles)-1]
	ar.LatestTRC = trcFile
	return nil
}

func (ar *CaApiServer) LoadClientsAndSecrets() error {
	for _, client := range ar.CaConfig.Clients {
		parts := strings.Split(client, ":")
		if len(parts) != 2 {
			return fmt.Errorf("Invalid client configuration: %s", client)
		}
		secretFile := filepath.Join(ar.CA.ConfigDir, parts[1])
		secretKey := caconfig.NewPEMSymmetricKey(secretFile)
		log.Println("[CA] Loaded secret for client ", parts[0])
		ar.Clients = append(ar.Clients, CaClient{
			ClientId: parts[0],
			Secret:   secretKey,
		})
	}

	return nil
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func (ar *CaApiServer) renewCert(wr http.ResponseWriter, req *http.Request) {

	_, err := ar.authJwt(req.Header.Get("authorization"))
	if err != nil {
		log.Println("[CA] JWT auth failed for token " + req.Header.Get("authorization"))
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "JWT auth failed", http.StatusUnauthorized)
		return
	}

	var renewRequest models.RenewalRequest
	if err := json.NewDecoder(req.Body).Decode(&renewRequest); err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Could not parse JSON request body", http.StatusBadRequest)
		return
	}
	isdNumber := chi.URLParam(req, "isdNumber")
	asNumber := chi.URLParam(req, "asNumber")
	log.Println("[CA] Got isd ", isdNumber, "and AS ", asNumber)

	file, err := os.CreateTemp("/tmp/", "*.csr")
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	bts, err := base64.StdEncoding.DecodeString(renewRequest.Csr)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ar.loadLatestTRC()
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csr, err := ExtractAndVerifyCsr(ar.LatestTRC, bts, file)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write csr to file in pem encoding
	/*csrPem := pem.EncodeToMemory(&pem.Block{

		Type:  "CERTIFICATE REQUEST",
		Bytes: bts,
	})
	csrFileName := filepath.Join("/tmp", fmt.Sprintf("%s.csr", randomString(16)))
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(csrFileName, csrPem, 0777)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}*/
	/*stepCli := step.NewStepCliAdapter()*/

	// certFile, err := os.CreateTemp("/tmp/", "*.crt")
	certFileName := filepath.Join("/tmp", fmt.Sprintf("%s.cert", randomString(16)))
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = os.Chmod(file.Name(), 0777)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	isdAS := fmt.Sprintf("%s-%s", isdNumber, asNumber)
	if len(csr.Subject.ExtraNames) > 0 {
		str, ok := csr.Subject.ExtraNames[0].Value.(string)
		if ok {
			isdAS = str
		}
	}
	log.Println("[CA] Got ISDAS ", isdAS)
	err = ar.CA.IssueCertificateFromCSR(file.Name(), certFileName, isdNumber, asNumber)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	/*err = stepCli.SignCert(file.Name(), certFileName, ar.CertDuration, isdAS)
	// os.Remove(file.Name())
	if err != nil {
		log.Println("[CA] Renew failed with error ",err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}*/

	respCertChain, err := ExtractCerts(certFileName)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// os.Remove(certFileName)

	resp := models.RenewalResponse{
		CertificateChain: respCertChain,
	}
	err = json.NewEncoder(wr).Encode(&resp)
	if err != nil {
		log.Println("[CA] Renew failed with error ", err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Could not write response", http.StatusInternalServerError)
		return
	}
}

func (ar *CaApiServer) authJwt(tokenStr string) (*jwt.Token, error) {

	// Remove bearer things
	realToken := strings.Replace(tokenStr, "Bearer ", "", 1)
	log.Println("[CA] Got token for auth ", realToken)

	timeOffset := ""
	timeOffsetEnv := os.Getenv("JWT_SUPPORTED_TIME_OFFSET_MINS")
	if timeOffsetEnv != "" {
		timeOffset = timeOffsetEnv
	}

	if timeOffset != "" {
		realTimeOffset, err := strconv.Atoi(timeOffset)
		if err != nil {
			log.Println("[CA] Failed to pase JWT_SUPPORTED_TIME_OFFSET_MINS=", timeOffsetEnv)
		} else {
			log.Println("[CA] Adjusting time offset to ", realTimeOffset, " minutes")
			jwt.TimeFunc = func() time.Time {
				return time.Now().Add(time.Minute * time.Duration(realTimeOffset))
			}
		}

	}

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	var token *jwt.Token
	var err error
	success := false
	for _, client := range ar.Clients {

		if success {
			break
		}

		// Try all clients
		func(client *CaClient) {
			token, err = jwt.Parse(realToken, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return client.Secret.Get() // TODO: Figure out real client here
			})
			if err == nil {
				success = true
			}
		}(&client)
	}

	if !success {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, nil
	} else {
		return nil, fmt.Errorf("Token invalid")
	}
}

func healthCheck(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Add("Cache-Control", "no-store")
	healthCheckStatus := models.HealthCheckStatus{
		Status: "available",
	}
	if err := json.NewEncoder(wr).Encode(&healthCheckStatus); err != nil {
		sendProblem(wr, "healthCheck", "Could not write Response", http.StatusInternalServerError)
		return
	}
}

func sendProblem(wr http.ResponseWriter, errorType, title string, status int32) {
	wr.WriteHeader(int(status))
	problem := models.Problem{
		Status: status,
		Type:   errorType,
		Title:  title,
	}
	_ = json.NewEncoder(wr).Encode(&problem)
}

func (ar *CaApiServer) Run() error {
	// TODO: TLS
	serverUrl := ":3000"
	if ar.CaConfig.Server != "" {
		serverUrl = ar.CaConfig.Server
	}

	log.Println("[CA] Starting CA API server on ", serverUrl)
	return http.ListenAndServe(serverUrl, ar.Router)
}
