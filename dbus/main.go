package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

func main() {
	//Connect DBus System bus
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Error connecting to DBus: ", err)
	}

	s, err := os.Open("./sdp_record.xml")
	if err != nil {
		log.Fatal("Error opening SDP XML", err)
	}

	sdp, err := ioutil.ReadAll(s)
	if err != nil {
		log.Fatal("Error reading SDP XML", err)
	}

	opts := map[string]dbus.Variant{
		"AutoConnect":   dbus.MakeVariant(true),
		"ServiceRecord": dbus.MakeVariant(bytes.NewBuffer(sdp).String()),
	}

	uuid := uuid.NewString()

	dbusChannel := make(chan *dbus.Call, 1)
	manager := conn.Object("org.bluez", "/org/bluez")
	register := manager.Go("org.bluez.ProfileManager1.RegisterProfile", 0, dbusChannel, (dbus.ObjectPath)("/stenobox/profile"), uuid, opts)

	if register.Err != nil {
		log.Fatal(register.Err)
	}

}
