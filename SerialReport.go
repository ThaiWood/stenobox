package main

import (
	"github.com/tarm/serial"
	"log"
	"strconv"
)

type SerialReport struct {
	Device string
	Baud   int
	SerialKeys
}

type SerialKeys struct {
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
}

func (r *SerialReport) Empty() {
	r.Mods = 0
}

func (r *SerialReport) SetKey(keycode int) {
	r.Key1 = keycode
}
func (r *SerialReport) SendKeys(port string) error {

	c := &serial.Config{Name: r.Device, Baud: r.Baud}

	serialPort, err := serial.OpenPort(c)
	defer serialPort.Close()

	if err != nil {
		log.Println("Error opening communications port or device, check permissions")
		return err
	}

	serialPort.Write([]byte("-"))

	serialPort.Write([]byte(strconv.Itoa(r.Mods)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key1)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key2)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key3)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key4)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key5)))
	serialPort.Write([]byte(","))
	serialPort.Write([]byte(strconv.Itoa(r.Key6)))

	serialPort.Write([]byte(";"))

	return nil
}

func (r *SerialReport) Setup() {
}
