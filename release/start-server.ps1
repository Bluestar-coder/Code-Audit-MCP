$ErrorActionPreference = "Stop"
$here = Split-Path -Parent $MyInvocation.MyCommand.Definition
cd $here
Write-Host "Starting CodeAuditMcp backend..." -ForegroundColor Green
./server.exe