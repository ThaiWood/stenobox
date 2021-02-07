module github.com/thaiwood/stenobox

go 1.15

require (
	github.com/godbus/dbus/v5 v5.0.3
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/thaiwood/stenobox/bluetooth v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
)

replace github.com/thaiwood/stenobox/bluetooth => ./bluetooth
