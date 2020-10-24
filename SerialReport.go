package main

import (
	"github.com/tarm/serial"
	"log"
	"strconv"
)

type SerialReport struct {
	Mods     int
	reserved int
	Key1     int
	Key2     int
	Key3     int
	Key4     int
	Key5     int
	Key6     int
}

func (r *SerialReport) AddModifier(keycode int) {
	r.Mods = keycode
	return
}

func (r *SerialReport) Empty() {
	r.Mods = 0
}

func (r *SerialReport) SetKey(keycode int) {
	r.Key1 = keycode
}
func (r *SerialReport) SendKeys(port string) error {

	c := &serial.Config{Name: "/dev/serial0", Baud: 115200}

	hid, err := serial.OpenPort(c)
	if err != nil {
		log.Println("Error opening communications port or device, check permissions")
		return err
	}

	hid.Write([]byte("-"))

	hid.Write([]byte(strconv.Itoa(r.Mods)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key1)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key2)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key3)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key4)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key5)))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(r.Key6)))

	hid.Write([]byte(";"))

	return nil
}
