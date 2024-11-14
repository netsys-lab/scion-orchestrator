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



build_and_copy_binaries() {
    cd dev/scion/
    local bin_dir="$1"

    # Ensure bin directory exists
    mkdir -p "../../$bin_dir"

    # Array of relative paths for each command
    declare -a commands=(
        "scion/cmd/scion"
        "scion-pki/cmd/scion-pki"
        "router/cmd/router"
        "control/cmd/control"
        "daemon/cmd/daemon"
    #    "dispatcher/cmd/dispatcher"
    )

    echo $(pwd)
    # Build each command with the specified environment variables
    for cmd_path in "${commands[@]}"; do
        cd "$cmd_path" || exit
        CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build
        cd - > /dev/null || exit
        echo $(pwd)
    done

   

    # Copy binaries to bin directory
    cp scion/cmd/scion/scion "../../$bin_dir/"
    cp scion-pki/cmd/scion-pki/scion-pki "../../$bin_dir/"
    cp router/cmd/router/router "../../$bin_dir/"
    cp control/cmd/control/control "../../$bin_dir/"
    cp daemon/cmd/daemon/daemon "../../$bin_dir/"
    #cp dispatcher/cmd/dispatcher/dispatcher "../../$bin_dir/"

    cd ../../

    # Final build with environment variables, if needed
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build

   
    # Copy final binary to bin directory
    if [ -f "scion-orchestrator.exe" ]; then
        cp scion-orchestrator.exe "$bin_dir/"
    else 
        cp scion-orchestrator "$bin_dir/"
    fi
}

#GOOS=linux GOARCH=amd64 build_and_copy_binaries "./bin_release/linux_amd64"
#GOOS=linux GOARCH=arm64 build_and_copy_binaries "./bin_release/linux_arm64"
#GOOS=darwin GOARCH=amd64 build_and_copy_binaries "./bin_release/darwin_amd64"
#GOOS=darwin GOARCH=arm64 build_and_copy_binaries "./bin_release/darwin_arm64"
GOOS=windows GOARCH=amd64 build_and_copy_binaries "./bin_release/windows_amd64"
