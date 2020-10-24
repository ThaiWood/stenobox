#include <Keyboard.h>
#include <SoftwareSerial.h>

//#define HWSERIAL Serial1
const byte rxPin = 9;
const byte txPin = 10;

SoftwareSerial HWSERIAL(rxPin, txPin);

const byte numChars = 32;
char receivedChars[numChars];

boolean newData = false;

void setup() {
  HWSERIAL.begin(19200);
  Serial.begin(19200);
}

void loop() {
  recvWithStartEndMarkers();
  showNewData();
}

void recvWithStartEndMarkers() {
  static boolean recvInProgress = false;
  static byte ndx = 0;
  char startMarker = '-';
  char endMarker = ';';
  char rc;

  while (HWSERIAL.available() > 0 && newData == false) {
    rc = HWSERIAL.read();

    if (recvInProgress == true) {
      if (rc != endMarker) {
        receivedChars[ndx] = rc;
        ndx++;
        if (ndx >= numChars) {
          ndx = numChars - 1;
        }
      }
      else {
        receivedChars[ndx] = '\0'; // terminate the string
        recvInProgress = false;
        ndx = 0;
        newData = true;
      }
    }

    else if (rc == startMarker) {
      recvInProgress = true;
    }
  }
}

void showNewData() {
  if (newData == true) {
    int keys[6];
    char *pos;
    newData = false;
    
    Serial.println(receivedChars);

    pos = strtok(receivedChars, ",");
    int i = 0;
    while (pos != NULL) {
      keys[i] = atoi(pos);
      if (keys[i] != 0 && i > 0) {
        // The built-in arduino library assumes everyting below 128 is printable and if aboe that, will subtract 136 and then send a scancode (which is what we want)
        keys[i] += 136;
      }
      pos = strtok(NULL, ",");
      i++;
    }

    if (keys[0] != 0) {
      switch (keys[0]) {
        case 0xE0:
          keys[0] = KEY_LEFT_CTRL;
          break;
        case 0xE1:
          keys[0] = KEY_LEFT_SHIFT;
          break;
        case 0xE2:
          keys[0] = KEY_LEFT_ALT;
          break;
        case 0xE3:
          keys[0] = KEY_LEFT_GUI;
      }
    }


    Keyboard.press(keys[0]);
    Keyboard.press(keys[1]);
    Keyboard.releaseAll();
  }
}
