param(
    [string]$OutputPath = ".\\gpresult.xml"
)

$dir = Split-Path -Path $OutputPath -Parent
if ($dir -and -not (Test-Path $dir)) {
    New-Item -ItemType Directory -Path $dir | Out-Null
}

Write-Host "Collecting gpresult XML to $OutputPath"
gpresult /X $OutputPath /F | Out-Null
Write-Host "Done"

