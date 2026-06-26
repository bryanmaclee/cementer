# claim — peter

> Advisory, **optimistic** (NOT a lock). Overwrite THIS file only; reset to idle at session close.

```
status:     active
session:    P4 (2026-06-25)
arc:        serial-split-tap (BUILD) — Intellisense channel first
branch:     peter/p2-serial-split-build
intent-sha: — (no source commits yet; hardware build + scope-doc capture)
since:      2026-06-25
note:       Resumed the serial-split build. Operator returned with measurement #1 for BOTH DAQs (Intellisense -6.35V idle / pin1=GND,pin2=TXD,transmit-only; Totco -8.2V idle / pin5=GND,pin2=TXD). Building the Intellisense single-channel listen tap first (6N137, Rin~1k, read 19200 8N1). NEW FINDING: Totco TX is DTR-gated (streams only while consumer asserts DTR/pin4) -> Totco validates in coexistence, not Pi-only. scope.md update + measurements pending. peter/p3-doc-currency doc PR still pending operator auth.
```
