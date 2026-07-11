$ErrorActionPreference = 'Stop'
$utf8 = New-Object System.Text.UTF8Encoding($false, $true)
$extensions = @('.md','.json','.yaml','.yml','.go','.ts','.tsx','.vue','.sql','.ps1','.toml')
$errors = @()
Get-ChildItem -Recurse -File | Where-Object {
  $extensions -contains $_.Extension.ToLowerInvariant() -and
  $_.FullName -notmatch '[\\/]node_modules[\\/]'
} | ForEach-Object {
  try {
    $bytes = [IO.File]::ReadAllBytes($_.FullName)
    $text = $utf8.GetString($bytes)
    if ($text.Contains([char]0xFFFD)) { $errors += "replacement character: $($_.FullName)" }
  } catch { $errors += "invalid UTF-8: $($_.FullName)" }
}
if ($errors.Count -gt 0) { $errors | ForEach-Object { Write-Error $_ }; exit 1 }
Write-Output 'encoding verification passed'
