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
    git checkout f0d570b1cdf7cfd374abb5efc91aa68cc489ee0d
    cd ../../
fi
cd dev/scion/

echo "Building SCION for Windows"
cd scion/cmd/scion
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
 cd ../../../scion-pki/cmd/scion-pki
 echo "Building SCION pki for Windows"
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
 echo "Building SCION Router for Windows"
 cd ../../../router/cmd/router
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
 echo "Building SCION Control for Windows"
 cd ../../../control/cmd/control
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
 echo "Building SCION Daemon for Windows"
 cd ../../../daemon/cmd/daemon
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
cd ../../..

cd ../../

mkdir -p bin_windows

cp dev/scion/scion/cmd/scion/scion.exe ./bin_windows/
cp dev/scion/scion-pki/cmd/scion-pki/scion-pki.exe ./bin_windows/
cp dev/scion/router/cmd/router/router.exe ./bin_windows/
cp dev/scion/control/cmd/control/control.exe ./bin_windows/
cp dev/scion/daemon/cmd/daemon/daemon.exe ./bin_windows/
