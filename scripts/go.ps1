$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoDir = Split-Path -Parent $ScriptDir
$RequiredGo = (Get-Content -Raw (Join-Path $RepoDir "toolchain/go-version.txt")).Trim()
$GoCommand = if ($env:GO) { $env:GO } else { "go" }

Remove-Item Env:GOROOT -ErrorAction SilentlyContinue
$env:GOTOOLCHAIN = $RequiredGo
$env:GOEXPERIMENT = ""

$ActualGo = (& $GoCommand env GOVERSION).Trim()
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
if ($ActualGo -ne $RequiredGo) {
    Write-Error "Go toolchain mismatch: required $RequiredGo, got $ActualGo"
    exit 2
}

& $GoCommand @args
exit $LASTEXITCODE
