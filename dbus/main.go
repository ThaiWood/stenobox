package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/godbus/dbus"

	"golang.org/x/sys/unix"
	"syscall"
)

const (
	CONTROLPSM   = 0x11
	INTERRUPTPSM = 0x13
)

type DBusMessageReceiver struct {
	path        dbus.ObjectPath
	InterruptFD int
	ControlFD   int
}

func NewDBusMessageReceiver(path string) *DBusMessageReceiver {
	return &DBusMessageReceiver{
		path: (dbus.ObjectPath)(path),
	}
}

func (d *DBusMessageReceiver) Path() dbus.ObjectPath {
	return d.path
}

func (d *DBusMessageReceiver) Close() {
	log.Println("Receiving closing message, nothing to do")
}

func (d *DBusMessageReceiver) NewConnection(dev dbus.ObjectPath, fd dbus.UnixFD, fdProps map[string]dbus.Variant) *dbus.Error {
	fmt.Printf("New connection: %+v, %+v, %+v\n", dev, fd, fdProps)

	/*
		infd, isa, err := unix.Accept(d.InterruptFD)
		if err != nil {
			unix.Close(d.InterruptFD)
			log.Fatal("Error accpeting connection on interrupt socket", err)
		}
		log.Printf("ACCEPTED CONNECTION: infd: %d, isa %+v\n", infd, isa)

		// TODO: Now do something with the incoming fd??
	*/
	return nil

}

func MakeControlSocket() int {
	controlSocket, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_SEQPACKET, unix.BTPROTO_L2CAP)

	if err = syscall.SetsockoptInt(controlSocket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}

	err = unix.Bind(controlSocket, &unix.SockaddrL2{
		Addr: [6]uint8{0, 0, 0, 0, 0},
		PSM:  CONTROLPSM,
	})
	if err != nil {
		log.Fatal("Failed to bind control socket: ", err)
	}

	if err := unix.Listen(controlSocket, 5); err != nil {
		log.Fatal("Error listening on control socket", err)
	}

	return controlSocket

}

func MakeInterruptSocket() int {

	interruptSocket, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_SEQPACKET, unix.BTPROTO_L2CAP)
	if err != nil {
		log.Fatal("Error creating interrupt socket", err)
	}

	if err = syscall.SetsockoptInt(interruptSocket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}

	log.Println("Attemping to bind sockets....")

	err = unix.Bind(interruptSocket, &unix.SockaddrL2{
		Addr: [6]uint8{0, 0, 0, 0, 0},
		PSM:  INTERRUPTPSM,
	})
	if err != nil {
		log.Fatal("Failed to bind interrupt  socket: ", err)
	}

	log.Println("Interrupt socket bound")

	if err := unix.Listen(interruptSocket, 5); err != nil {
		log.Fatal("Error listening on interrupt socket", err)
	}

	//	cnfd, csa, err := unix.Accept(controlSocket)
	//	if err != nil {
	//		log.Fatal("Error accpeting connection on control socket", err)
	//	}
	//	log.Printf("cnfd: %d, csa %+v\n", cnfd, csa)
	//
	//
	//	log.Println("Listening.....")

	return interruptSocket
}

func (d *DBusMessageReceiver) Release() *dbus.Error {
	return nil
}

//type foo string

//func (f foo) Foo() (string, *dbus.Error) {
//	fmt.Println(f)
//	return string(f), nil
//}

func main() {
	log.Println("Setting up bluetooth")
	setupBluetooth()
	log.Println("Bluetooth setup")

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Error connecting to DBus: ", err)
	}

	d := NewDBusMessageReceiver("/thaiwood/stenobox/profile")

	if err := conn.Export(d, d.Path(), "org.bluez.Profile1"); err != nil {
		log.Fatal(err)
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
		"PSM":         dbus.MakeVariant(uint16(INTERRUPTPSM)),
		"AutoConnect": dbus.MakeVariant(true),
		//"RequireAuthentication": dbus.MakeVariant(false),
		//"RequireAuthorization":  dbus.MakeVariant(false),
		"ServiceRecord": dbus.MakeVariant(bytes.NewBuffer(sdp).String()),
	}

	uuid := "00001124-0000-1000-8000-00805f9b34fb"

	dbusChannel := make(chan *dbus.Call, 1)
	manager := conn.Object("org.bluez", "/org/bluez")
	register := manager.Go("org.bluez.ProfileManager1.RegisterProfile", 0, dbusChannel, d.Path(), uuid, opts)

	if register.Err != nil {
		log.Fatal(register.Err)
	}

	d.InterruptFD = MakeInterruptSocket()
	//d.ControlFD = MakeControlSocket()

	nfd, sa, err := unix.Accept(d.InterruptFD)
	if err != nil {
		log.Fatal("Error accpeting connection on interrupt socket", err)
	}
	log.Printf("INTERRUPT ACCEPT: nfd: %d, sa %+v\n", nfd, sa)

	d.InterruptFD = nfd

	//nfd, sa, err = unix.Accept(d.ControlFD)
	//if err != nil {
	//	log.Fatal("Error accepting connection on control socket", err)
	//}
	//log.Printf("CONTROL ACCEPT: nfd: %d, sa %+v\n", nfd, sa)

	//d.ControlFD = nfd

	//time.Sleep(30 * time.Second)

	//_, err = unix.Write(d.ControlFD, []byte{0x2, 0, 0, 0xb, 0, 0, 0, 0, 0})
	//if err != nil {
	//	log.Println("Failed to write to control socket", err)
	//}

	//log.Println("Wrote Data")

	//_, err = unix.Write(d.InterruptFD, []byte{0xa1, 0x13, 0x03})
	//if err != nil {
	//	log.Println("Failed to write to interrupt  socket", err)
	//}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	evloop := true
	for evloop {
		select {
		case dbusObjectCall := <-dbusChannel:
			if dbusObjectCall.Err != nil {
				log.Fatal(err)
				evloop = false
			}
		case <-sig:
			log.Println("Will quit")
			evloop = false
		default:
		}
	}

	log.Println("Trying to unregister profile")
	unregister := manager.Call("org.bluez.ProfileManager1.UnregisterProfile", 0, d.Path())
	if unregister.Err != nil {
		log.Println("Failed to unregister profile: ", unregister.Err)
	}

	close(dbusChannel)
	conn.Close()

}

func MakeOurService(*dbus.Conn) error {

	return nil
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
