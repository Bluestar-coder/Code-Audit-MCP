param(
    [switch]$VerboseOutput
)
$ErrorActionPreference = "Stop"
$root = Resolve-Path (Get-Location)
$repoRoot = $root.Path
$script:issues = @()

function AddIssue($type, $path) {
    $obj = [pscustomobject]@{ Type = $type; Path = $path }
    $script:issues += $obj
}
function UnderRelease($fullPath) {
    try {
        $rel = [System.IO.Path]::GetFullPath($fullPath)
        $releaseRoot = [System.IO.Path]::Combine($repoRoot, "release")
        return $rel.StartsWith($releaseRoot, [System.StringComparison]::OrdinalIgnoreCase)
    } catch { return $false }
}

# Known directories to check (relative to repo root)
$dirCandidates = @(
    "tests",
    "frontend\build",
    "frontend\dist",
    "frontend\coverage",
    "frontend\node_modules\.cache",
    "proto\__pycache__",
    "backend\proto\__pycache__",
    "backend\bin"
)
foreach($rel in $dirCandidates) {
    $full = Join-Path $repoRoot $rel
    if (Test-Path -LiteralPath $full) {
        if (-not (UnderRelease $full)) { AddIssue "dir" $full }
    }
}

# Generic cache dirs anywhere
Get-ChildItem -Path $repoRoot -Directory -Recurse -ErrorAction SilentlyContinue |
  Where-Object { $_.Name -in @("__pycache__", ".pytest_cache", ".mypy_cache") -and -not (UnderRelease $_.FullName) } |
  ForEach-Object { AddIssue "cache-dir" $_.FullName }

# Stray .pyc files anywhere
Get-ChildItem -Path $repoRoot -Recurse -File -Include *.pyc -ErrorAction SilentlyContinue |
  Where-Object { -not (UnderRelease $_.FullName) } |
  ForEach-Object { AddIssue "pyc" $_.FullName }

# Dev binary (should not exist in repo root)
$devExe = Join-Path $repoRoot "backend\server.exe"
if (Test-Path -LiteralPath $devExe) { AddIssue "dev-exe" $devExe }

if ($script:issues.Count -eq 0) {
    Write-Host "Clean state verified: no tests, caches, or dev artifacts." -ForegroundColor Green
    exit 0
} else {
    Write-Host "Found unwanted files/directories:" -ForegroundColor Red
    $script:issues | ForEach-Object { Write-Host ("[{0}] {1}" -f $_.Type, $_.Path) }
    if ($VerboseOutput) {
        Write-Host "\nHint: run cleanup, or add exceptions if absolutely necessary (not recommended)." -ForegroundColor Yellow
    }
    exit 1
}