$ErrorActionPreference = 'Stop'
$matrix = 'docs/traceability/requirements-matrix.md'
if (-not (Test-Path -LiteralPath $matrix)) { throw 'requirements matrix missing' }
$content = Get-Content -LiteralPath $matrix -Raw -Encoding UTF8
if ($content -notmatch 'PRD/Decision' -or $content -notmatch 'Evidence') { throw 'requirements matrix schema incomplete' }
$legacy = Get-ChildItem -LiteralPath 'docs/legacy/openspec-0.17/changes' -Directory | Where-Object { $_.Name -match '^\d{3}-' }
$map = Get-Content -LiteralPath 'docs/traceability/legacy-change-map.md' -Raw -Encoding UTF8
$missing = @($legacy | Where-Object { $map -notmatch [regex]::Escape($_.Name.Substring(0,3)) })
if ($missing.Count -gt 0) { throw "legacy mappings missing: $($missing.Name -join ', ')" }
$legacyRoot = (Resolve-Path 'docs/legacy/openspec-0.17').Path
$manifestPath = Join-Path $legacyRoot 'MANIFEST.sha256'
if (-not (Test-Path -LiteralPath $manifestPath)) { throw 'legacy manifest missing' }
$verified = 0
Get-Content -LiteralPath $manifestPath -Encoding UTF8 | Where-Object { $_ -match '^[0-9a-f]{64}  ' } | ForEach-Object {
  $expected, $relative = $_ -split '  ', 2
  $target = Join-Path $legacyRoot ($relative -replace '/', '\')
  if (-not (Test-Path -LiteralPath $target)) { throw "legacy file missing: $relative" }
  $actual = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne $expected) { throw "legacy hash mismatch: $relative" }
  $verified++
}
if ($verified -eq 0) { throw 'legacy manifest contains no hashes' }
Write-Output "traceability verification passed: $($legacy.Count) changes mapped, $verified legacy files verified"
