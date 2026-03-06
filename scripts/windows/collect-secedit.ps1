param(
    [string]$OutputPath = ".\\secedit.inf"
)

$dir = Split-Path -Path $OutputPath -Parent
if ($dir -and -not (Test-Path $dir)) {
    New-Item -ItemType Directory -Path $dir | Out-Null
}

Write-Host "Collecting secedit INF to $OutputPath"
secedit /export /cfg $OutputPath | Out-Null
Write-Host "Done"

