import serial
import time

port = "COM3"
baudrate = 115200

with serial.Serial(port, baudrate=baudrate, timeout=1) as ser:
    time.sleep(2)
    with open("EnbridgeCC4-16CICR@3250.csv", "r", encoding="utf-8-sig") as f:
        for i, line in enumerate(f):
            clean = line.strip()
            ser.write((clean + "\n").encode("utf-8"))
            print(f"Line {i}: '{clean}'")
            time.sleep(0.05)            