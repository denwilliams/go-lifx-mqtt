package lifx

import "go.yhsif.com/lifxlan"

func toPowerPayload(power lifxlan.Power) bool {
	if power == 0 {
		return false
	}
	return true
}
