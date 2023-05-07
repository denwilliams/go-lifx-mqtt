package light

import (
	"go.yhsif.com/lifxlan"
)

// Wrap tries to wrap a lifxlan.Device into a light device.
func Wrap(d lifxlan.Device) Device {

	ld := &device{
		Device: d,
	}

	return ld
}
