# intellisense-send.ps1 -- emit Intellisense 14-field frames out an RS-232 COM port.
#
# Bench/proto transmitter for the serial-split tap. Pairs with the opto front-end:
# real RS-232 path (Waveshare USB->RS232), Rin = 1k, NO inversion -- the adapter
# drives true RS-232 (mark negative) and the opto un-inverts it at the Pi.
#
# Wire contract (S4 capture, intellisense-wire-capture-2026-06-16.md):
#   19200 8N1, CR/LF, ~1 line/s, 14 comma fields, headerless.
#   Col 0 = HH:MM:SS uptime (resets on boot). Cols emit a slow triangle-wave
#   pump cycle so the live chart moves.
#
# Find the COM#:  [System.IO.Ports.SerialPort]::GetPortNames()
# Usage:          .\intellisense-send.ps1 -Port COM6
# If blocked by execution policy:
#   powershell -ExecutionPolicy Bypass -File .\intellisense-send.ps1 -Port COM6

param([string]$Port = 'COM6', [int]$Baud = 19200)

$sp = New-Object System.IO.Ports.SerialPort
$sp.PortName = $Port
$sp.BaudRate = $Baud
$sp.Parity   = [System.IO.Ports.Parity]::None
$sp.DataBits = 8
$sp.StopBits = [System.IO.Ports.StopBits]::One
$sp.Open()
Write-Host "Sending Intellisense frames on $Port @ $Baud 8N1 (Ctrl+C to stop)"

$volJob = 0.0; $volStage = 0.0; $t = 0
$inv = [System.Globalization.CultureInfo]::InvariantCulture   # force '.' decimal -> never commas in the CSV
$fmt = '{0:00}:{1:00}:{2:00},{3:F2},{4},{5:F2},{6:F1},{7},{8},{9:F2},{10:F2},{11:F1},{12:F2},{13:F1},{14:F1},{15}'

try {
  while ($true) {
    $hh = [int]($t / 3600) % 100; $mm = [int]($t / 60) % 60; $ss = $t % 60
    $phase = ($t % 120) / 120.0
    $ramp  = if ($phase -lt 0.5) { $phase * 2.0 } else { (1.0 - $phase) * 2.0 }

    $rate     = 4.6 * $ramp
    $pressure = [int][math]::Round(1306.0 * $ramp)
    $density  = 8.20 + 0.02 * $ramp
    $volJob   += $rate / 60.0; $volStage += $rate / 60.0

    # cols: 0 logtime,1 density.1,2 agg.pressure,3 agg.rate,4 vol.job,5 unit1.pressure,
    #       6 unit2.pressure,7 unit1.rate,8 unit2.rate,9 water.rate,10 density.2,
    #       11 vol.water.stage,12 vol.stage,13 job.number
    $line = [string]::Format($inv, $fmt,
      $hh, $mm, $ss, $density, $pressure, $rate, $volJob, $pressure, 0, $rate, 0.0, 0.0, 0.0, 0.0, $volStage, 0)

    $sp.Write($line + "`r`n")     # CR/LF, matches the real wire
    Write-Host $line
    Start-Sleep -Seconds 1
    $t++
  }
} finally { $sp.Close(); Write-Host "Port closed." }
