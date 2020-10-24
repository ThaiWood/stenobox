package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

type HIDReport struct {
	Device string
	HIDKeys
}

type HIDKeys struct {
	Mods     uint8
	reserved uint8
	Key1     uint8
	Key2     uint8
	Key3     uint8
	Key4     uint8
	Key5     uint8
	Key6     uint8
}

func (r *HIDReport) AddModifier(keycode int) {
	mods := map[int]byte{
		224: 0x01,
		225: 0x02,
		226: 0x04,
		227: 0x08,
	}

	if r.Mods == 0 {
		r.Mods = mods[keycode]
		return
	} else {
		r.Mods = r.Mods | mods[keycode]
	}
}

func (r *HIDReport) Empty() {
	r.Mods = 0
}

func (r *HIDReport) SendKeys(device string) error {
	release := HIDReport{}

	buf := new(bytes.Buffer)
	releasebuf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, r.HIDKeys)
	if err != nil {
		log.Println("Error writing HID report buffer")
		log.Println(err)
	}

	hid, err := os.OpenFile(device, os.O_WRONLY, 0666)
	defer hid.Close()
	if err != nil {
		return err
	}

	err = binary.Write(releasebuf, binary.LittleEndian, release.HIDKeys)
	if err != nil {
		log.Println("Error writing release HID report buffer")
		log.Print(err)
	}

	if _, err := hid.Write(buf.Bytes()); err != nil {
		hid.Close() // ignore error; Write error takes precedence
		log.Println("Error writing to " + device)
		log.Fatal(err)
	}

	if _, err := hid.Write(releasebuf.Bytes()); err != nil {
		hid.Close() // ignore error; Write error takes precedence
		log.Println("Error writing to " + device)
		log.Fatal(err)
	}

	return nil
}

func (r *HIDReport) Close() {
}

func (r *HIDReport) SetKey(hidcode int) {
	r.HIDKeys.Key1 = uint8(hidcode)
}
