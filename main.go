package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type Reporter interface {
	SendKeys(string) error
	SetKey(int)
	Empty()
	AddModifier(int)
	Setup()
}

const start = '-'
const end = ';'

var protocol string
var device string
var baud int

func init() {
	flag.StringVar(&protocol, "out", "bluetooth", "Output method, either serial or usb")
	flag.StringVar(&device, "dev", "/dev/serial0", "Output device, such as /dev/serial0 or /dev/hidg0")
	flag.IntVar(&baud, "baud", 115200, "Baud rate")
	flag.Parse()
}

func main() {

	var r Reporter

	switch protocol {
	case "serial":
		r = &SerialReport{Device: device, Baud: baud}
	case "usb":
		r = &HIDReport{}
	case "bluetooth":
		r = &BluetoothReport{}
		r.Setup()
	default:
		log.Fatal("Protocol must either be serial, usb, or bluetooth")
	}

	rDown, _ := regexp.Compile(`key press\s*(?P<code>\d*)`)
	//rUp := regexp.Compile(`key release\s(.*\d)`)

	s := setupXinput()

	for s.Scan() {
		fmt.Println(s.Text())
		match := rDown.FindStringSubmatch(s.Text())

		if len(match) > 0 {
			keycode, err := strconv.Atoi(match[1])

			if err != nil {
				log.Println("Error converting string to integer")
				log.Print(err)
				continue
			}

			hidcode := XorgToHID(keycode)

			if hidcode == -1 {
				continue
			}

			if hidcode > 223 {
				r.AddModifier(hidcode)
				fmt.Printf("ADD MODIFIER: %+v\n", r)
			} else {

				r.SetKey(int(hidcode))

				err = r.SendKeys(device)

				if err != nil {
					log.Println("Error sending keys to recieving device: ")
					log.Println(err)
				}

				r.Empty()

				//	hid.Close()
			}

		}
	}

}

// Adapated from https://gist.github.com/precondition/cdf18eadc2a9f5600311a17ef58e5f45

func HID_to_Xorg(keycode int) int {
	return ((getHIDKeyboard())[keycode] + 8)
}

func XorgToHID(keycode int) int {
	i := findKeyPosition(keycode - 8)
	return i
}

func findKeyPosition(keycode int) int {
	keyboard := getHIDKeyboard()
	for i, _ := range keyboard {
		if keyboard[i] == keycode {
			return i
		}
	}
	return -1
}

func getHIDKeyboard() []int {
	return []int{
		0, 0, 0, 0, 30, 48, 46, 32, 18, 33, 34, 35, 23, 36, 37, 38,
		50, 49, 24, 25, 16, 19, 31, 20, 22, 47, 17, 45, 21, 44, 2, 3,
		4, 5, 6, 7, 8, 9, 10, 11, 28, 1, 14, 15, 57, 12, 13, 26,
		27, 43, 43, 39, 40, 41, 51, 52, 53, 58, 59, 60, 61, 62, 63, 64,
		65, 66, 67, 68, 87, 88, 99, 70, 119, 110, 102, 104, 111, 107, 109, 106,
		105, 108, 103, 69, 98, 55, 74, 78, 96, 79, 80, 81, 75, 76, 77, 71,
		72, 73, 82, 83, 86, 127, 116, 117, 183, 184, 185, 186, 187, 188, 189, 190,
		191, 192, 193, 194, 134, 138, 130, 132, 128, 129, 131, 137, 133, 135, 136, 113,
		115, 114, -1, -1, -1, 121, -1, 89, 93, 124, 92, 94, 95, -1, -1, -1,
		122, 123, 90, 91, 85, -1, -1, -1, -1, -1, -1, -1, 111, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, 179, 180, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, 111, -1, -1, -1, -1, -1, -1, -1,
		29, 42, 56, 125, 97, 54, 100, 126, 164, 166, 165, 163, 161, 115, 114, 113,
		150, 158, 159, 128, 136, 177, 178, 176, 142, 152, 173, 140, -1, -1, -1, -1}
}

func setupXinput() *bufio.Scanner {
	xinputcmd := "xinput"
	args := []string{"test", "Virtual core XTEST keyboard"}

	xinput := exec.Command(xinputcmd, args...)
	xinput.Env = append(os.Environ(), "DISPLAY=:0")

	xinputReader, err := xinput.StdoutPipe()
	if err != nil {
		fmt.Println("Error setting up xinput reader: ")
		log.Fatal(err)
	}

	err = xinput.Start()
	if err != nil {
		log.Println("Error running xinput command")
		log.Fatal(err)
	}

	return bufio.NewScanner(xinputReader)
}

func setupBluetooth() {
	hciconfigcmd := "hciconfig"
	args := []string{"hci0", "class", "0x0025C0"}
	hciconfig := exec.Command(hciconfigcmd, args...)

	err := hciconfig.Start()
	if err != nil {
		log.Println("Error setting hci class:")
		log.Fatal(err)
	}

	args = []string{"hci0", "name", "StenoBox"}
	hciconfig = exec.Command(hciconfigcmd, args...)

	err = hciconfig.Start()
	if err != nil {
		log.Println("Error setting hci name:")
		log.Fatal(err)
	}

	args = []string{"hci0", "piscan"}
	hciconfig = exec.Command(hciconfigcmd, args...)

	err = hciconfig.Start()
	if err != nil {
		log.Println("Error making device discoverable va hciconfig")
		log.Fatal(err)
	}

}
