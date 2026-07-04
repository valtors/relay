$ErrorActionPreference = "Stop"
Set-Location "C:\Users\91876\Downloads\AI CMO\relay"
$env:ANTHROPIC_API_KEY = "dummy-test-key"

# JSON-RPC frames over stdio
$initReq = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"verify","version":"0.1"}}}'
$initNote = '{"jsonrpc":"2.0","method":"notifications/initialized"}'
$listReq  = '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

$input = "$initReq`n$initNote`n$listReq`n"

$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = (Resolve-Path ".\relay.exe").Path
$psi.WorkingDirectory = (Get-Location).Path
$psi.RedirectStandardInput  = $true
$psi.RedirectStandardOutput = $true
$psi.RedirectStandardError  = $true
$psi.UseShellExecute = $false
$psi.CreateNoWindow  = $true

$p = [System.Diagnostics.Process]::Start($psi)
$p.StandardInput.Write($input)
$p.StandardInput.Close()
$stdout = $p.StandardOutput.ReadToEnd()
$stderr = $p.StandardError.ReadToEnd()
$p.WaitForExit(3000) | Out-Null
if (-not $p.HasExited) { $p.Kill(); $p.WaitForExit(1000) | Out-Null }

Write-Host "===== STDERR =====" -ForegroundColor Cyan
Write-Host $stderr
Write-Host "`n===== STDOUT (raw JSON-RPC frames) =====" -ForegroundColor Cyan
Write-Host $stdout

Write-Host "`n===== TOOL NAMES =====" -ForegroundColor Green
$lines = $stdout -split "`n" | Where-Object { $_.Trim().StartsWith("{") }
foreach ($line in $lines) {
  try {
    $obj = $line | ConvertFrom-Json
    if ($obj.id -eq 2 -and $obj.result.tools) {
      $i = 0
      foreach ($t in $obj.result.tools) {
        $i++
        Write-Host ("  {0}. {1}" -f $i, $t.name)
      }
      Write-Host ("`n  TOTAL: {0} tools registered" -f $obj.result.tools.Count)
    }
  } catch { }
}
