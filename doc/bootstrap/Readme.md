# Bootstrapping Server
Per default, `scion-orchestrator` runs a bootstrapping server on all IPs on the address `:8041`. Endhosts can bootstrap into the AS by configuring the proper IP of the AS host followed by the port `8041`. If the endhost is running the `scion-orchestrator` tool, this will happen automatically if the `bootstrap.server` is configured accordingly.

Following configurations are possible in the `scion-orchestrator.toml`:

```toml
[bootstrap]
# Setting the bootstrap server URL for the AS host to listen on, or the endhost to connect to
server = "127.0.0.1:8041"

# Overwrite fields in the topology e.g. for endhosts who have NAT between them and the routers.
# This can be done by accessing fields in the topology.json via `.` and overwrite them with different values
# If this is configured, these updates are stored in a `topology_endhost.json` file, so the actual topology is not affected
# Only endhosts will get this changed topology.
topology_overwrites = [
    "border_routers.br-1.internal_addr=127.0.0.1:30004",
    "control_service.cs-1.addr=127.0.0.1:34003",
]

# This property allows to define subnets from which requests to the bootstrapping server are allowed
# If empty, all source IPs are allowed.
# If this list contains any entries, all requests from other source IPs are served with 403 forbidden
allowed_subnets = [
    "10.0.0.0/8",
    "192.168.0.0/24",
]
```
