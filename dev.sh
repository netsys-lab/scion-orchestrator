# !/bin/bash
set -e
# Check if dev directory exists
if [ -d "dev" ]; then
    echo "Directory dev exists."
else 
    mkdir dev
    cd dev
    git clone https://github.com/scionproto/scion.git
    cd scion
    git checkout v0.12.0
    cd ../../
fi
cd dev/scion/


cd scion/cmd/scion
 CGO_ENABLED=0 go build
 cd ../../../scion-pki/cmd/scion-pki
 CGO_ENABLED=0 go build
 cd ../../../router/cmd/router
 CGO_ENABLED=0 go build
 cd ../../../control/cmd/control
 CGO_ENABLED=0 go build
 cd ../../../daemon/cmd/daemon
 CGO_ENABLED=0 go build
 cd ../../../dispatcher/cmd/dispatcher
 CGO_ENABLED=0 go build
cd ../../..

cd ../../

mkdir -p bin

cp dev/scion/scion/cmd/scion/scion ./bin/
cp dev/scion/scion-pki/cmd/scion-pki/scion-pki ./bin/
cp dev/scion/router/cmd/router/router ./bin/
cp dev/scion/control/cmd/control/control ./bin/
cp dev/scion/daemon/cmd/daemon/daemon ./bin/
cp dev/scion/dispatcher/cmd/dispatcher/dispatcher ./bin/

CGO_ENABLED=0 go build