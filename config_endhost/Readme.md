# Sample Endhost configuration to run an Endhost within the OVGU AS (71-2:0:4a)

## Get the binaries
To build the tooling, you need to have `go >= 1.23` and `git` installed. 

### MacOS / Linux
Start the build process by running `./dev.sh`. This will clone the latest tested SCION commit into `dev/` and build all the binaries for your platform and copy them into the `bin` directory.

### Windows (not fully operable yet)
Start the build process by running `dev.bat`. This will clone the latest tested SCION commit into `dev/` and build all the binaries for your platform and copy them into the `bin` directory.

## Build the scion-orchestrator tool
Run `go build` to build the actual `scion-orchestrator` tool.

## Run an example
Choose one of the examples and copy the content of the example folder into a `./config` folder, e.g. to run a full core AS including CA do the following:
```sh
mkdir -p config
cp -R examples/endhost-ovgu/* ./config/
sudo ./scion-orchestrator run 
```

Then open a second terminal and run `./bin/scion showpaths 71-20965`. This should print one path to this AS.

Please note the using `scion-orchestrator run` does not update your installed `scion` binary (under `/usr/bin/scion`), so for now please use the binaries under `./bin`.