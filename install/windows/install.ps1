# Temporary winget-shaped install until the package lands in winget-pkgs.
# Downloads the Windows release zip and puts qomp.exe on the user PATH.
$ErrorActionPreference = "Stop"
$ver = "1.0.2"
$url = "https://github.com/Jayanth1312/qompnt/releases/download/v$ver/qomp_windows_amd64.zip"
$dir = Join-Path $env:LOCALAPPDATA "qomp"
$zip = Join-Path $env:TEMP "qomp-$ver.zip"

New-Item -ItemType Directory -Force -Path $dir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $dir -Force
Remove-Item $zip -Force

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$dir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$dir", "User")
  $env:Path = "$env:Path;$dir"
}

Write-Host "Installed qomp $ver to $dir"
Write-Host "Open a new terminal, then run: qomp --version"
