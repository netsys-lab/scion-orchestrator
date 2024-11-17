param (
    [string]$Repository = "https://github.com/scionproto/scion.git",
    [string]$Tag = "v0.12.0",
    [string]$Path = "dev"
)
Set-StrictMode -Version Latest

function Clone-Scion {
    param (
        [string]$Repository = "https://github.com/scionproto/scion.git",
        [string]$Tag = "v0.12.0",
        [string]$Path
    )
    git clone --branch $Tag $Repository $Path
}

function Build-Scion {
    param (
        [string]$Path
    )
    $out = "$(Get-Location)\bin"
    $env:CGO_ENABLED = 0
    Push-Location $Path
    try {
        New-Item -Path $out -ItemType "directory" -Force | Out-Null
        go build -o $out .\control\cmd\control\
        Check-Error
        go build -o $out .\daemon\cmd\daemon\
        Check-Error
        # Dispatcher doesn't build on Windows
        #go build -o $out .\dispatcher\cmd\dispatcher\
        #Check-Error
        go build -o $out .\router\cmd\router\
        Check-Error
        go build -o $out .\scion\cmd\scion\
        Check-Error
        go build -o $out .\scion-pki\cmd\scion-pki\
        Check-Error
    } finally {
        Remove-Item Env:\CGO_ENABLED
        Pop-Location
    }
}

function Check-Error {
    if ($lastexitcode -ne 0) {
        throw "Go build failed"
    }
}

try {
    if (-Not (Test-Path -Path $Path)) {
        Clone-Scion -Path $Path -Repository $Repository -Tag $Tag
    }
    Build-SCION -Path $Path
    Copy-Item -Path bin -Destination .\integration -Recurse -Force
} catch {
    Write-Host "Error building SCION:"
    Write-Host $_
}
