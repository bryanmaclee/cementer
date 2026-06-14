<#
  serial-read.ps1 — direct-laptop serial capture / discovery for the DAQ raw feed.

  Reads a COM port for a bounded time, dumps hex + ASCII (with <CR>/<LF> markers),
  reports a printable-byte ratio (high ~ correct baud/framing), and saves the raw
  bytes to a .bin under -OutDir when any data arrives. No Pi / Go / Node needed.

  Usage:
    # normal read (Totco confirmed 9600 8N1, Protocol 1, 1/4-sec interval):
    powershell -File tools/serial-read.ps1 -Port COM6 -Baud 9600 -Seconds 10

    # baud sweep when the rate is unknown (silence at ALL bauds = physical, not settings):
    powershell -File tools/serial-read.ps1 -Port COM6 -Sweep

    # loopback self-test (jumper pin 2<->3 on the adapter DB9, unplugged from the DAQ):
    powershell -File tools/serial-read.ps1 -Port COM6 -Loopback

  Notes / gotchas learned 2026-06-14:
   - Total SILENCE (0 bytes) at every baud means no signal on RX -> electrical/physical:
     a straight-through cable between two DTE ends (needs a NULL-MODEM/crossover), or the
     DAQ not actually transmitting, or a dead adapter. Wrong baud gives GARBAGE, not silence.
   - DTR + RTS are asserted below; some RS-232 instruments stay mute until those are high.
   - CD/CTS/DSR all False = handshake pins not driven (3-wire cable, or DAQ not ready).
#>
param(
  [string]$Port = 'COM6',
  [int]$Baud = 9600,
  [int]$Seconds = 10,
  [int]$DataBits = 8,
  [System.IO.Ports.Parity]$Parity = [System.IO.Ports.Parity]::None,
  [System.IO.Ports.StopBits]$StopBits = [System.IO.Ports.StopBits]::One,
  [string]$OutDir = (Join-Path $PSScriptRoot '..\captures'),
  [switch]$Sweep,
  [switch]$Loopback
)

function Read-Port([int]$baud, [int]$secs) {
  $p = New-Object System.IO.Ports.SerialPort $Port, $baud, $Parity, $DataBits, $StopBits
  $p.Handshake = [System.IO.Ports.Handshake]::None
  $p.DtrEnable = $true; $p.RtsEnable = $true
  $p.ReadTimeout = 300; $p.WriteTimeout = 500
  try { $p.Open() } catch { Write-Output "OPEN FAILED on $Port @ $baud : $($_.Exception.Message)"; return $null }
  if ($Loopback) { $p.DiscardInBuffer(); $p.Write("CEMENTER-LOOPBACK-TEST-12345`r`n") }
  $buf = New-Object System.Collections.Generic.List[byte]
  $deadline = (Get-Date).AddSeconds($secs)
  while ((Get-Date) -lt $deadline) {
    try { $b = $p.ReadByte(); if ($b -ge 0) { $buf.Add([byte]$b) } }
    catch [System.TimeoutException] { } catch { break }
  }
  $lines = "CD:$($p.CDHolding) CTS:$($p.CtsHolding) DSR:$($p.DsrHolding)"
  $p.Close()
  [pscustomobject]@{ Bytes = $buf.ToArray(); Modem = $lines }
}

if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

if ($Sweep) {
  foreach ($b in 2400,4800,9600,19200,38400,57600,115200) {
    $r = Read-Port $b 2
    if ($null -eq $r) { continue }
    $n = $r.Bytes.Count
    Write-Output ("{0,6} {1}N{2} : {3,4} bytes   {4}" -f $b, $DataBits, $StopBits.value__, $n, $r.Modem)
  }
  Write-Output "(0 bytes at every baud = physical/electrical, not a settings problem)"
  return
}

$r = Read-Port $Baud $Seconds
if ($null -eq $r) { return }
$bytes = $r.Bytes; $n = $bytes.Count
Write-Output "=== $Port @ $Baud ${DataBits}N$($StopBits.value__) : $n bytes in ${Seconds}s ==="
Write-Output "Modem lines -> $($r.Modem)"
if ($Loopback) {
  $txt = -join ($bytes | ForEach-Object { if ($_ -ge 32 -and $_ -le 126){[char]$_} else {'.'} })
  if ($txt -match 'CEMENTER-LOOPBACK-TEST-12345') { Write-Output "LOOPBACK OK -> adapter + port + software GOOD; fault is the DAQ-side path." }
  else { Write-Output "LOOPBACK FAILED ($n bytes) -> not jumpered, or adapter/driver fault." }
  return
}
if ($n -eq 0) { Write-Output "No data. See script header: silence = physical, not settings."; return }
$printable = ($bytes | Where-Object { ($_ -ge 32 -and $_ -le 126) -or $_ -eq 9 -or $_ -eq 10 -or $_ -eq 13 }).Count
Write-Output ("Printable ratio: {0}%" -f [math]::Round(100.0*$printable/$n,1))
$show = [Math]::Min($n,500)
Write-Output "`n--- HEX (first $show) ---"; Write-Output (($bytes[0..($show-1)] | ForEach-Object { $_.ToString('X2') }) -join ' ')
$ascii = -join ($bytes[0..($show-1)] | ForEach-Object {
  if ($_ -eq 13){'<CR>'} elseif ($_ -eq 10){"<LF>`n"} elseif ($_ -eq 9){'<TAB>'} elseif ($_ -ge 32 -and $_ -le 126){[char]$_} else {'.'} })
Write-Output "`n--- ASCII ---"; Write-Output $ascii
$stamp = Get-Date -Format 'yyyy-MM-ddTHHmmss'
$out = Join-Path $OutDir ("capture-$stamp-$Baud-${DataBits}N$($StopBits.value__).bin")
[System.IO.File]::WriteAllBytes($out, $bytes)
Write-Output "`nSaved: $out"
