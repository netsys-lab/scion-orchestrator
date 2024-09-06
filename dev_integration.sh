# !/bin/bash
set -e

# Check if dev directory exists
if [ -d "dev" ]; then
    echo "Directory dev exists."
    echo "Deleting dev directory."
    rm -rf dev
    mkdir dev
    cd dev
    git clone https://github.com/scionproto/scion.git
    cd scion
    git checkout f51e6dab7
    cd ../../
fi
cd dev/scion/

cd scion/cmd/scion
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 cd ../../../scion-pki/cmd/scion-pki
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 cd ../../../router/cmd/router
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 cd ../../../control/cmd/control
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 cd ../../../daemon/cmd/daemon
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 cd ../../../dispatcher/cmd/dispatcher
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
cd ../../..

cd ../../

mkdir -p integration/bin

cp dev/scion/scion/cmd/scion/scion ./integration/bin/
cp dev/scion/scion-pki/cmd/scion-pki/scion-pki ./integration/bin/
cp dev/scion/router/cmd/router/router ./integration/bin/
cp dev/scion/control/cmd/control/control ./integration/bin/
cp dev/scion/daemon/cmd/daemon/daemon ./integration/bin/
cp dev/scion/dispatcher/cmd/dispatcher/dispatcher ./integration/bin/
