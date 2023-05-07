package lifx

import "go.yhsif.com/lifxlan"

type LIFXType uint16

const (
	Unknown LIFXType = 0
	Light   LIFXType = 100
	Switch  LIFXType = 200
)

func getType(hw *lifxlan.HardwareVersion) (LIFXType, *lifxlan.Product) {
	if hw == nil {
		return Unknown, nil
	}

	if hw.VendorID != 1 {
		return Unknown, nil
	}

	key := lifxlan.ProductMapKey(hw.VendorID, hw.ProductID)
	product := lifxlan.ProductMap[key]

	hasRelays := product.Features.Relays

	// This is a bit of a lazy way to do this
	if hasRelays != nil && *hasRelays {
		return Switch, &product
	}

	// Could do more here, but for now, just assume it's a light.
	// Might want to separate tiles and candles
	return Light, &product
}
