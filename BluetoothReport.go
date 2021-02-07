package main

import (
	"github.com/thaiwood/stenobox/bluetooth"
)

type BluetoothReport struct {
	Device string
	Baud   int
	BluetoothKeys
	DBus bluetooth.DBusMessageReceiver
}

type BluetoothKeys struct {
	Mods     uint8
	reserved uint8
	Key1     uint8
	Key2     uint8
	Key3     uint8
	Key4     uint8
	Key5     uint8
	Key6     uint8
}

func (r *BluetoothReport) AddModifier(hidcode int) {
	mods := map[int]byte{
		224: 0x01,
		225: 0x02,
		226: 0x04,
		227: 0x08,
	}

	if r.Mods == 0 {
		r.Mods = mods[hidcode]
	} else {
		r.Mods = r.Mods | mods[hidcode]
	}
}

func (r *BluetoothReport) Empty() {
	r.Mods = 0
}

func (r *BluetoothReport) SetKey(hidcode int) {
	r.Key1 = uint8(hidcode)
}

func (r *BluetoothReport) makeByteSlice() []byte {
	return []byte{0xA1, 1, byte(r.Mods), 0, byte(r.Key1), byte(r.Key2), byte(r.Key3), byte(r.Key4), byte(r.Key5), byte(r.Key6)}
}

func (r *BluetoothReport) SendKeys(port string) error {

	r.DBus.InterruptCh <- r.makeByteSlice()

	return nil
}

func (r *BluetoothReport) Setup() {
	r.DBus = *bluetooth.NewDBusMessageReceiver("/thaiwood/stenobox/profile")
	go func() { r.DBus.StartBluetooth() }()
}
