# scion-orchestrator: Run a cross-platform SCION Host (Endhost or AS) 
This tool called `scion-orchestrator` allows to setup SCION connectivity on a host, either as a regular process (`standalone`) or as installed service. Depending on the configuration it can run SCION infrastructure components (`Control Service`, `Border Router`, ...) to form a full AS or a regular `endhost` stack. `scion-orchestrator` is designed to support both cases and let the hosts running the SCION AS provide bootstrapping servers that the endhost can use to become part of the SCION AS.

So far we do not provide pre-built binaries, so please take a look at **Get the binaries** to build all the tooling

## Supported Platforms
Standalone/run:
- [x] MacOS Arm64
- [x] MacOS Amd64
- [x] Linux Amd64
- [x] Windows Amd64

Install:
- [?] MacOS Arm64
- [?] MacOS Amd64
- [?] Linux Amd64
- [?] Windows Amd64

## Usage
`scion-orchestrator` requires two directories in the current working directory:
- `./bin`: Contains built SCION binaries (daemon, dispatcher, router, etc)
- `./config`: Contains AS/Host configurations (see examples)

On Linux/MacOS:
```sh
# Start all configured SCION components as standalone processes
./scion-orchestrator run

# Install SCION components as system services
./scion-orchestrator install

# Shutdown all installed SCION system services
./scion-orchestrator shutdown
```

On Windows:
```sh
# Start all configured SCION components as processes
scion-orchestrator.exe run

# Install SCION components as system services
scion-orchestrator.exe install

# Shutdown all installed SCION system services
scion-orchestrator.exe shutdown
```

## Get the binaries
To build the tooling, you need to have `go >= 1.23` and `git` installed. 

### MacOS / Linux
Start the build process by running `./dev.sh`. This will clone the latest tested SCION commit into `dev/` and build all the binaries for your platform and copy them into the `bin` directory.

### Windows
For Windows, we don't have an automated build script yet, so please follow the next commands (in minGW or Powershell), or try `dev.bat` but without warranty so far.
```bat
mkdir dev
git clone https://github.com/scionproto/scion.git
cd scion
git checkout v0.12.0

# Build windows binaries 
go build -o ../../bin ./router/... ./control/... ./dispatcher/... ./daemon/... ./scion/... ./scion-pki/... ./gateway/...
cd ../../

# Build orchestrator
go build
```

## Run an example
Choose one of the examples and copy the content of the example folder into a `./config` folder, e.g. to run a full core AS including CA do the following:
```sh
mkdir -p config
cp -R examples/core-as-ca/* ./config/
sudo ./scion-orchestrator run # scion-orchestrator.exe on windows
```
**Note: Ensure you have the binaries under `./bin`.**

## Run as AS host

### Metrics / AS Status

### Certificate Renewal
AS hosts that participate in the control plane (e.g. by running a `Control Service`) can perform a renewal of the AS certificate. `scion-orchestrator` does this periodically in the background by checking if the AS certificate expires soon and performing a renew when a configured threshold is passed. 

**Future Work**:
- Offline Renewal: Provide ASes a key to issue tokens to connect to a central instance via regular IP allowing to issue a new certificate when SCION connectivity to the CA is not given.
- API to work with certificates: Perform check, listing and renewal operations via API

### CA 
Issuing core ASes can use `scion-orchestrator` to run a dedicated CA to issue SCION AS certificates. To achieve this, configure the following fields in the `scion-orchestrator.toml` to run a CA Server:

```toml
[ca]
server = ":3000"
clients = ["123:client1.secret"]
```

The `server` field configures the HTTP endpoint to run the CA HTTP API. The `clients` allows to set one or more API clients, which are usually `Control Service` instances. Each client has a clientId and a secret, a symmetric key. The secret is configured as a filepath, starting at the root config directory of `scion-orchestrator`. The client needs to issue `JSON Web Tokens` with the given secret to authenticate at the CA server.

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
Per default, `scion-orchestrator` runs a bootstrapping server on all IPs on the address `:8041`. Endhosts can bootstrap into the AS by configuring the proper IP of the AS host followed by the port `8041`. If the endhost is running the `scion-orchestrator` tool, this will happen automatically if the `bootstrap.server` is configured accordingly.

## Run as Endhost
`scion-orchestrator` offers to set the `mode` in the `scion-orchestrator.toml` to `endhost`. This will force the host to fetch the SCION configuration from the address configured in `bootstrap.server`. In `endhost` mode the host is still able to run the `Metrics Server` configured via `metrics.server`. The endhost mode requires to connect to a valid bootstrapping URL to start and a responding bootstrapping server.

A minimal configuration for an endhost that does not run any infrastructure SCION components is:

```toml
isd_as = "1-150"
mode = "endhost"

[metrics]
server = "127.0.0.1:33401"

[bootstrap]
server = "10.150.0.254:8041"
```

## Integration Testing
At first, build the `scion-orchestrator` tool and copy the binary into the `integration` folder.
```sh 
go build && cp scion-orchestrator ./integration
```

Then download the latest SCION binary release from [github](https://github.com/scionproto/scion/releases/tag/v0.11.0) and copy all binaries into `integration/bin/`.

Then run `docker compose up -d` to start the ASes `150` and `151` and the endhost in `150`.