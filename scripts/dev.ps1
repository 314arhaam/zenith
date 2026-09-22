# Optional wrapper. Direct invocation via py -3 scripts/dev.py also works.
$ErrorActionPreference = 'Stop'
$scriptPath = Join-Path $PSScriptRoot 'dev.py'
if ($env:PYTHON) {
    & $env:PYTHON $scriptPath @args
} elseif (Get-Command py -ErrorAction SilentlyContinue) {
    & py -3 $scriptPath @args
} elseif (Get-Command python -ErrorAction SilentlyContinue) {
    & python $scriptPath @args
} else {
    throw 'Python 3.10+ is required for developer tools, not for the Zenith binaries.'
}
exit $LASTEXITCODE
