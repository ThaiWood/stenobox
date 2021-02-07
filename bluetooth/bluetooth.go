package bluetooth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

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
	ControlCh   chan byte
	InterruptCh chan []byte
}

func NewDBusMessageReceiver(path string) *DBusMessageReceiver {
	return &DBusMessageReceiver{
		path:        (dbus.ObjectPath)(path),
		ControlCh:   make(chan byte, 2),
		InterruptCh: make(chan []byte),
	}
}

func (d *DBusMessageReceiver) Path() dbus.ObjectPath {
	return d.path
}

func (d *DBusMessageReceiver) Close() {
	log.Println("Receiving closing message, nothing to do")
}

func (d *DBusMessageReceiver) NewConnection(dev dbus.ObjectPath, fd dbus.UnixFD, fdProps map[string]dbus.Variant) *dbus.Error {
	fmt.Printf("New connection: DEV: %+v, FD: %+v, FDPROPS: %+v\n", dev, fd, fdProps)
	fmt.Printf("Our Data:\n\tInterrupt: %+v\n\tControl: %+v\n", d.InterruptFD, d.ControlFD)

	nfd, _, _ := AcceptInterrupt(d.InterruptFD)
	d.InterruptFD = nfd

	sa, err := unix.Getsockname(int(fd))
	if err != nil {
		log.Printf("Couldn't get sock name from FD %d: %s", int(fd), err)
	}
	log.Printf("Got new SA: %+v", sa)

	d.ControlFD = int(fd)
	/*
		err = syscall.SetNonblock(d.ControlFD, false)
		if err != nil {
			log.Println("Failed to set control socket to blocking", err)
		} */

	log.Println("Sending hello on ctrl channel")
	if _, err := unix.Write(d.ControlFD, []byte{0xa1, 0x13, 0x03}); err != nil {
		log.Println("Failure on Sending Hello on Ctrl 1", err)
	}

	if _, err := unix.Write(d.ControlFD, []byte{0xa1, 0x13, 0x02}); err != nil {
		log.Println("Failure on Sending Hello on Ctrl 2", err)
	}

	time.Sleep(5 * time.Second)

	message := []byte{0xA1, 0x01, 0, 0, 4, 0, 0, 0, 0, 0}

	if _, err := unix.Write(d.InterruptFD, message); err != nil {
		log.Println("Failed to write HID message", err)
	}

	time.Sleep(10 * time.Millisecond)

	message = []byte{0xA1, 0x01, 0, 0, 0, 0, 0, 0, 0, 0}

	if _, err := unix.Write(d.InterruptFD, message); err != nil {
		log.Println("Failed to write HID message", err)
	}

	go d.Loop()
	return nil

}

func (d *DBusMessageReceiver) WriteHID(hid []byte) {
	if _, err := unix.Write(d.InterruptFD, hid); err != nil {
		log.Println("Error writing hid data: ", err)
	}
	time.Sleep(10 * time.Millisecond)
	release := []byte{0xA1, 0x01, 0, 0, 0, 0, 0, 0, 0, 0}
	if _, err := unix.Write(d.InterruptFD, release); err != nil {
		log.Println("Error writing release data: ", err)
	}
}

func (d *DBusMessageReceiver) Loop() {

	for {
		select {
		case <-d.ControlCh:
			log.Println("Got data on control channel, quitting handler")
			return
		case msg := <-d.InterruptCh:
			log.Printf("Got data from appliaction: %+v\n", msg)
			d.WriteHID(msg)
		default:
			/*
				buf := make([]byte, 1024)
				datasize, err := unix.Read(d.ControlFD, buf)
				if err != nil || datasize < 1 {
					log.Println("no data received - quitting event loop", err)
					d.Close()
					return
				}
				log.Printf("Data from socket: %+v\n", buf)
			*/
		}
	}
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
	log.Println("Bound control socket")

	if err := unix.Listen(controlSocket, 1); err != nil {
		log.Fatal("Error listening on control socket", err)
	}
	log.Println("Listening on control socket")

	return controlSocket

}

func AcceptControl(fd int) (int, unix.Sockaddr, error) {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		log.Printf("\n Error accepting control connecton on FD: %d: %s", fd, err)
	} else {
		log.Printf("Accepted control and got: %d and %+v\n", nfd, sa)
	}
	return nfd, sa, err
}

func AcceptInterrupt(fd int) (int, unix.Sockaddr, error) {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		log.Printf("\n Error accepting interrupt connecton on FD: %d: %s", fd, err)
	} else {
		log.Printf("Accepted interrupt  and got: %d and %+v\n", nfd, sa)
	}
	return nfd, sa, err
}

func MakeInterruptSocket() int {

	interruptSocket, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_SEQPACKET, unix.BTPROTO_L2CAP)
	if err != nil {
		log.Fatal("Error creating interrupt socket", err)
	}

	if err = syscall.SetsockoptInt(interruptSocket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}

	log.Println("Attemping to bind interrupt socket")

	err = unix.Bind(interruptSocket, &unix.SockaddrL2{
		Addr: [6]uint8{0, 0, 0, 0, 0},
		PSM:  INTERRUPTPSM,
	})
	if err != nil {
		log.Println("Failed to bind interrupt  socket: ", err)
	}

	log.Println("Interrupt socket bound")

	if err := unix.Listen(interruptSocket, 1); err != nil {
		log.Fatal("Error listening on interrupt socket", err)
	}
	log.Println("Listening on interrupt socket")

	//
	//
	//	log.Println("Listening.....")

	return interruptSocket
}

func (d *DBusMessageReceiver) RequestDisconnection(dev dbus.ObjectPath) *dbus.Error {
	log.Printf("Disconnection requested from: %+v\n", dev)

	//	err := unix.Close(d.ControlFD)
	//	if err != nil {
	//		log.Println("Failed to close control socket during disconnection request", err)
	//	}
	//
	//	err = unix.Close(d.InterruptFD)
	//	if err != nil {
	//		log.Println("Failed to close control socket during disconnection request", err)
	//	}

	return nil
}

func (d *DBusMessageReceiver) Release() *dbus.Error {
	return nil
}

func (d *DBusMessageReceiver) StartBluetooth() {
	log.Println("Setting up bluetooth")
	setupBluetooth()
	log.Println("Bluetooth setup")

	log.Println("Making Interrupt  socket")
	d.InterruptFD = MakeInterruptSocket()

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Error connecting to DBus: ", err)
	}

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
		"PSM":         dbus.MakeVariant(uint16(CONTROLPSM)),
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
