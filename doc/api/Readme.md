# SCION-Orchestrator API Documentation

The scion-orchestrator provides a REST API served via HTTPS and secured via basic auth. It can be configured under the `api` section in the `scion-orchestrator.toml`:

```toml
[api]
users = ["admin:admin.secret"] # user:password
address = ":8843" # Default, served via HTTPS
```

The API is secured by custom certificate, so curl commands need to be executed with `-k` to ignore a missing root certificate.

## Infra-Host handler
**Note: When running in mode == "as", these handlers are available via the API.**

### POST /api/v1/cppki/csr
Create a Certificate Signing Request from the host-locals private key.

#### Example
```bash
curl -k -u "admin:admin.secret" -X POST https://localhost:8443/api/v1/cppki/csr -d @csr.json
```

#### Input
```json
{
    "subject": {
        "common_name": "1-150 AS Certificate",
        "isd_as": "1-150"
    }
}
```

#### Result
PEM-encoded CSR as response body.

### POST /api/v1/cppki/certs
Upload a new certificate chain for the control plane.

#### Example
```bash
curl -k -u "admin:admin.secret" -X POST https://localhost:8443/api/v1/cppki/certs -d @cert.pem
```

#### Input
PEM-encoded certificate chain.

#### Result
```json
{
    "message": "Certificate chain added successfully"
}
```

## CA Handler
**Note: If you have configured the CA section and have a proper CA certificate on the host, the following endpoints are available.**

### POST ca/certs/:isd/:as/sign
Sign a certificate signing request with the configured CA certificate to issue a SCION control plane certificate.

#### Example
```bash
curl -k -u "admin:admin.secret" -X POST https://localhost:8443/api/v1/ca/certs/1/150/sign -d @as.csr
```

#### Input
PEM-encoded certificate signing request (e.g. generated via `/api/v1/cppki/csr`).

#### Result
PEM-encoded certificate chain.
