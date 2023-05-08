package lifx

import (
	"math"

	"go.yhsif.com/lifxlan"
)

func isSameColor(a *lifxlan.Color, b *lifxlan.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.Hue == b.Hue &&
		a.Saturation == b.Saturation &&
		a.Brightness == b.Brightness &&
		a.Kelvin == b.Kelvin
}

type colorPayload struct {
	Hue        uint8  `json:"hue"`
	Saturation uint8  `json:"saturation"`
	Brightness uint8  `json:"brightness"`
	Kelvin     uint16 `json:"kelvin"`
}

func toColorPayload(color *lifxlan.Color) *colorPayload {
	if color == nil {
		return nil
	}

	return &colorPayload{
		Hue:        uint16to8(color.Hue),
		Saturation: uint16to8(color.Saturation),
		Brightness: uint16toPercent(color.Brightness),
		Kelvin:     color.Kelvin,
	}
}

func uint16to8(value uint16) uint8 {
	return uint8(value >> 8)
}

func uint16toPercent(value uint16) uint8 {
	return uint8(math.Round(float64(value) / math.MaxUint16 * 100))
}
