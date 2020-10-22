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
	mods_const := 0xE000
	mods := map[int]int{
		224: 0x01,
		225: 0x02,
		226: 0x04,
		227: 0x08, //META_L
	}

	r.Mods = mods[keycode] | mods_const
	return
}

func (r *SerialReport) Empty() {
	r.Mods = 0
}

func (r *SerialReport) SetKey(keycode int) {

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
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key1))))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key2))))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key3))))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key4))))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key5))))
	hid.Write([]byte(","))
	hid.Write([]byte(strconv.Itoa(makeKeyForSerial(r.Key6))))

	hid.Write([]byte(";"))

	return nil
}

// Key data from: https://github.com/PaulStoffregen/cores/blob/master/teensy4/keylayouts.h

func makeKeyForSerial(keycode int) int {
	bytes := map[int]int{
		0x81: 0xE200,
		0x82: 0xE200,
		0x83: 0xE200,
		0xB0: 0xE400,
		0xB1: 0xE400,
		0xB2: 0xE400,
		0xB3: 0xE400,
		0xB4: 0xE400,
		0xB5: 0xE400,
		0xB6: 0xE400,
		0xB7: 0xE400,
		0xB8: 0xE400,
		0xB9: 0xE400,
		0xCD: 0xE400,
		0xCE: 0xE400,
		0xE2: 0xE400,
		0xE9: 0xE400,
		0xEA: 0xE400,
		4:    0xF000,
		5:    0xF000,
		6:    0xF000,
		7:    0xF000,
		8:    0xF000,
		9:    0xF000,
		10:   0xF000,
		11:   0xF000,
		12:   0xF000,
		13:   0xF000,
		14:   0xF000,
		15:   0xF000,
		16:   0xF000,
		17:   0xF000,
		18:   0xF000,
		19:   0xF000,
		20:   0xF000,
		21:   0xF000,
		22:   0xF000,
		23:   0xF000,
		24:   0xF000,
		25:   0xF000,
		26:   0xF000,
		27:   0xF000,
		28:   0xF000,
		29:   0xF000,
		30:   0xF000,
		31:   0xF000,
		32:   0xF000,
		33:   0xF000,
		34:   0xF000,
		35:   0xF000,
		36:   0xF000,
		37:   0xF000,
		38:   0xF000,
		39:   0xF000,
		40:   0xF000,
		41:   0xF000,
		42:   0xF000,
		43:   0xF000,
		44:   0xF000,
		45:   0xF000,
		46:   0xF000,
		47:   0xF000,
		48:   0xF000,
		49:   0xF000,
		50:   0xF000,
		51:   0xF000,
		52:   0xF000,
		53:   0xF000,
		54:   0xF000,
		55:   0xF000,
		56:   0xF000,
		57:   0xF000,
		58:   0xF000,
		59:   0xF000,
		60:   0xF000,
		61:   0xF000,
		62:   0xF000,
		63:   0xF000,
		64:   0xF000,
		65:   0xF000,
		66:   0xF000,
		67:   0xF000,
		68:   0xF000,
		69:   0xF000,
		70:   0xF000,
		71:   0xF000,
		72:   0xF000,
		73:   0xF000,
		74:   0xF000,
		75:   0xF000,
		76:   0xF000,
		77:   0xF000,
		78:   0xF000,
		79:   0xF000,
		80:   0xF000,
		81:   0xF000,
		82:   0xF000,
		83:   0xF000,
		84:   0xF000,
		85:   0xF000,
		86:   0xF000,
		87:   0xF000,
		88:   0xF000,
		89:   0xF000,
		90:   0xF000,
		91:   0xF000,
		92:   0xF000,
		93:   0xF000,
		94:   0xF000,
		95:   0xF000,
		96:   0xF000,
		97:   0xF000,
		98:   0xF000,
		99:   0xF000,
		100:  0xF000,
		101:  0xF000,
		104:  0xF000,
		105:  0xF000,
		106:  0xF000,
		107:  0xF000,
		108:  0xF000,
		109:  0xF000,
		110:  0xF000,
		111:  0xF000,
		112:  0xF000,
		113:  0xF000,
		114:  0xF000,
		115:  0xF000,
	}

	return (keycode | (bytes[keycode]))
}
