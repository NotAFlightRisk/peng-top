# Installs the latest peng-top release: irm <this file's raw url> | iex
$ErrorActionPreference = 'Stop'

$repo = 'NotAFlightRisk/peng-top'
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$tag = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
$archive = "peng-top_${tag}_windows_$arch.zip"
$url = "https://github.com/$repo/releases/download/$tag"
$dir = "$env:LOCALAPPDATA\peng-top"
$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "peng-top-$tag") -Force

Write-Host "Downloading peng-top $tag for windows/$arch"
Invoke-WebRequest "$url/$archive" -OutFile "$tmp\$archive"
Invoke-WebRequest "$url/checksums.txt" -OutFile "$tmp\checksums.txt"

$expected = (Select-String -Path "$tmp\checksums.txt" -Pattern " $archive$").Line.Split(' ')[0]
$actual = (Get-FileHash "$tmp\$archive" -Algorithm SHA256).Hash
if ($actual -ne $expected) { throw "Checksum didn't match, not installing" }

Expand-Archive "$tmp\$archive" -DestinationPath $dir -Force
Remove-Item $tmp -Recurse -Force

$path = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($path -split ';') -notcontains $dir) {
  [Environment]::SetEnvironmentVariable('Path', "$path;$dir", 'User')
  Write-Host "Added $dir to your PATH, so open a new terminal first"
}
Write-Host "Installed to $dir\peng-top.exe - run it with: peng-top"
