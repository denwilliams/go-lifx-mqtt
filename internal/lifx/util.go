package lifx

func onOrOff(state bool) string {
	if state {
		return "on"
	}
	return "off"
}
