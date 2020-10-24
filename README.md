# StenoBox

This is a collection of software and linux system configurations intended for Raspberry Pi's to make them run plover, take in serial input over USB, and output text via USB HID.  This means that the receiving device  _does not_ have to be running Plover.

## How it works

StenoBox uses `xinput test` to read the output of Plover on the virtual X keyboard and translates this into HID reports.

There are two methods this can be used:

## 1. Software only (Pi4 only)

If you're using a Pi4, you can use software only mode.  This expects that you'll:

1. Connect your pi4's USB-C port to the device that should receive input
1. Create a hidgadget (see `starthid.sh` and [kernel docs](https://www.kernel.org/doc/Documentation/usb/gadget_configfs.txt) for more on gadgets)
1. Start plover
1. Start StenoBox

## 2. Hardware help mode

If you're using a Pi Zero (or some other device possibly), StenoBox will write keycodes via serial to an Arduino device that can then be used with the Arudino Keyboard library to write the hid reports.

This mode is compatible with a wider range of receiving hardware, such as (in my testing):

1. MacBook Pro (13-inch, 2017, Four Thunderbolt 3 Ports) running OSX 10.15.7
1. Acer C720 running Debian Buster
1. Ipad pro 11 running iPadOS 14

## What about the GUI?

Plover needs an X environment regardless of whether or not you pass `-g none`. This is what actually makes this software work too, we're able to grab the `xinput`.  

In order to run plover in a headless mode, we rely on the `dummy` xorg driver.  I've included an xorg.conf here that can be used.  It is of an OK screen size, that if you want to debug and do stuff you can run `x11vnc` or similar and get a decent screensize.  In the future I'll likely make the default screen much smaller if it saves much memory (untested as of yet).

Right now I start X manually just because I didn't get around to scripting it.  If you have your pi in GUI start mode you could probably just edit `.xinitrc` to do the startup.

## Under the hood

StenoBox is written in Go because it was easier for me, but ports to other languages are encouraged.  When the output is intended for an arudino type device (currently only tested with Teensy v4.1), StenoBox will produce keycodes generated from TeensyDuino library sources.

Also, hidgadget setup is thanks to [KoiOates](https://github.com/KoiOates/plover/tree/ploverducken).  For my personal setup, I'm using other gadget scripts, but these are a good start that focus only on HID, unlike many others that involve ethernet.

## In Progress

- Move more manual stuff to systemd, such as:
  - Starting X
  - Starting Plover
- Add more config flags and a better help
- Add a battery so that pi's can be unplugged/replugged without losing power
- Add a shutdown button so pi's can be shutdown before power removal

## Misc

There's also some udev rules in here just to make life easier, one to map a given keyboard's serial port to `/dev/ploverinput` so configs can stay the same regardless of device and one to ensure that hidgadgets that are created at `/dev/hidg*` are writeable by a user.
