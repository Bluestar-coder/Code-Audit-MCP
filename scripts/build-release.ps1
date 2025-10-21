param(
  [switch]$SkipBackend,
  [switch]$SkipFrontend,
  [switch]$VerboseOutput
)
$ErrorActionPreference = "Stop"
$root = Resolve-Path (Get-Location)
$repoRoot = $root.Path
$releaseDir = Join-Path $repoRoot "release"

function Ensure-Dir($p) { if (-not (Test-Path -LiteralPath $p)) { New-Item -ItemType Directory -Path $p | Out-Null } }
function Copy-Dir-Clean($src, $dst) {
  if (Test-Path -LiteralPath $dst) { Remove-Item -LiteralPath $dst -Force -Recurse }
  Ensure-Dir (Split-Path -Parent $dst)
  Copy-Item -Path $src -Destination $dst -Recurse
}

Ensure-Dir $releaseDir

# Build backend (Go)
if (-not $SkipBackend) {
  Write-Host "[backend] building server.exe" -ForegroundColor Cyan
  $backendDir = Join-Path $repoRoot "backend"
  if (-not (Test-Path -LiteralPath $backendDir)) { throw "backend directory not found: $backendDir" }
  Push-Location $backendDir
  try {
    go version | Out-Null
  } catch {
    throw "Go toolchain not found in PATH. Please install Go and retry."
  }
  $outExe = Join-Path $releaseDir "server.exe"
  $ldflags = "-s -w"
  # Prefer main package if present
  $hasCmdServer = Test-Path -LiteralPath (Join-Path $backendDir "cmd\server")
  if ($hasCmdServer) {
    go build -trimpath -ldflags $ldflags -o $outExe ./cmd/server
  } else {
    go build -trimpath -ldflags $ldflags -o $outExe .
  }
  Write-Host "[backend] output: $outExe" -ForegroundColor Green
  Pop-Location
}

# Build frontend (React)
if (-not $SkipFrontend) {
  Write-Host "[frontend] building static files" -ForegroundColor Cyan
  $frontendDir = Join-Path $repoRoot "frontend"
  if (-not (Test-Path -LiteralPath $frontendDir)) { throw "frontend directory not found: $frontendDir" }
  Push-Location $frontendDir
  try { npm -v | Out-Null } catch { throw "Node.js/npm not found in PATH. Please install Node and retry." }
  if (Test-Path -LiteralPath (Join-Path $frontendDir "package-lock.json")) {
    npm ci
  } else {
    npm install
  }
  npm run build
  Pop-Location
  $buildDir = Join-Path $frontendDir "build"
  if (-not (Test-Path -LiteralPath $buildDir)) { throw "frontend build output not found: $buildDir" }
  $releaseFrontend = Join-Path $releaseDir "frontend"
  Copy-Dir-Clean $buildDir $releaseFrontend
  Write-Host "[frontend] output: $releaseFrontend" -ForegroundColor Green
}

# Copy rules
Write-Host "[rules] syncing YAML from rules/common" -ForegroundColor Cyan
$rulesSrc = Join-Path $repoRoot "rules\common"
$rulesDst = Join-Path $releaseDir "rules\common"
Ensure-Dir (Split-Path -Parent $rulesDst)
if (Test-Path -LiteralPath $rulesSrc) {
  Copy-Dir-Clean $rulesSrc $rulesDst
  Write-Host "[rules] output: $rulesDst" -ForegroundColor Green
} else {
  Write-Warning "rules/common not found, skipping"
}

# Summary
Write-Host "\nBuild-release completed:" -ForegroundColor Green
Get-ChildItem -Path $releaseDir -Recurse | Select-Object FullName | ForEach-Object { Write-Host $_.FullName }