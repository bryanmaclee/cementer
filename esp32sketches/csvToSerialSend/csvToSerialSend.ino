#define RXD2 16
#define TXD2 17

void setup() {
  Serial.begin(115200);
  delay(500);  // let boot noise settle
  Serial2.begin(115200, SERIAL_8N1, RXD2, TXD2);
  delay(100);
  while(Serial.available()) Serial.read();  // flush buffer
  Serial.println("ESP32 Ready");
}

void loop() {
  if (Serial.available()) {
    String line = Serial.readStringUntil('\n');
    Serial2.println(line);
    Serial.println("Relayed: " + line);
  }
}