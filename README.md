# Scion-AS: Run a cross-platform SCION Host (Endhost or AS) 
This tool called `scion-as` allows to setup SCION connectivity on a host, either as a regular process (`standalone`) or as installed service. Depending on the configuration it can run SCION infrastructure components (`Control Service`, `Border Router`, ...) to form a full AS or a regular `endhost` stack. `scion-as` is designed to support both cases and let the hosts running the SCION AS provide bootstrapping servers that the endhost can use to become part of the SCION AS.

## Run as AS host

### Metrics / AS Status

### Certificate Renewal
AS hosts that participate in the control plane (e.g. by running a `Control Service`) can perform a renewal of the AS certificate. `scion-as` does this periodically in the background by checking if the AS certificate expires soon and performing a renew when a configured threshold is passed. 

**Future Work**:
- Offline Renewal: Provide ASes a key to issue tokens to connect to a central instance via regular IP allowing to issue a new certificate when SCION connectivity to the CA is not given.
- API to work with certificates: Perform check, listing and renewal operations via API

### CA 
Issuing core ASes can use `scion-as` to run a dedicated CA to issue SCION AS certificates. To achieve this, configure the following fields in the `scion-as.toml` to run a CA Server:

```toml
[ca]
server = ":3000"
clients = ["123:client1.secret"]
```

The `server` field configures the HTTP endpoint to run the CA HTTP API. The `clients` allows to set one or more API clients, which are usually `Control Service` instances. Each client has a clientId and a secret, a symmetric key. The secret is configured as a filepath, starting at the root config directory of `scion-as`. The client needs to issue `JSON Web Tokens` with the given secret to authenticate at the CA server.

The `Control Service` needs to have the following CA configuration:
```toml
[ca]
mode = "delegating"

[ca.service]
shared_secret = "/root/AS150/config/client1.secret" # Update to your location
addr = "http://127.0.0.1:3000"                      # Point to CA Server
client_id = "123"
```

**Future Work**:
- Force CA Server to use HTTPS and add the root cert to trust store

### Bootstrapping Server
Per default, `scion-as` runs a bootstrapping server on all IPs on the address `:8041`. Endhosts can bootstrap into the AS by configuring the proper IP of the AS host followed by the port `8041`. If the endhost is running the `scion-as` tool, this will happen automatically if the `bootstrap.server` is configured accordingly.

## Run as Endhost
`scion-as` offers to set the `mode` in the `scion-as.toml` to `endhost`. This will force the host to fetch the SCION configuration from the address configured in `bootstrap.server`. In `endhost` mode the host is still able to run the `Metrics Server` configured via `metrics.server`. The endhost mode requires to connect to a valid bootstrapping URL to start and a responding bootstrapping server.

A minimal configuration for an endhost that does not run any infrastructure SCION components is:

```toml
isd_as = "1-150"
mode = "endhost"

[metrics]
server = "127.0.0.1:33401"

[bootstrap]
server = "10.150.0.254:8041"
```