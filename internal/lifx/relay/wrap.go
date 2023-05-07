package relay

import (
	"go.yhsif.com/lifxlan"
)

// Wrap tries to wrap a lifxlan.Device into a relay device.
func Wrap(d lifxlan.Device) Device {
	if t, ok := d.(Device); ok {
		return t
	}

	return &device{
		Device: d,
	}
}
