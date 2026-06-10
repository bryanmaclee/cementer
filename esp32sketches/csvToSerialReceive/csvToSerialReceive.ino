#define RXD2 16
#define TXD2 17

void setup() {
  Serial.begin(115200);
  Serial2.begin(115200, SERIAL_8N1, RXD2, TXD2);
  delay(100);
  while(Serial2.available()) Serial2.read();  // flush boot noise
  Serial.println("ESP32 #2 Ready - Listening...");
}

void loop() {
  if (Serial2.available()) {
    String line = Serial2.readStringUntil('\n');
    line.trim();
    
    // only print if line contains a comma (valid CSV)
    // and only printable ASCII characters
    if (line.indexOf(',') != -1 && line.length() > 0) {
      bool valid = true;
      for (int i = 0; i < line.length(); i++) {
        if (line[i] < 32 || line[i] > 126) {
          valid = false;
          break;
        }
      }
      if (valid) Serial.println(line);
    }
  }
}