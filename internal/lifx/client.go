package lifx

import (
	"io"
	"math"
	"strings"

	"github.com/2tvenom/golifx"
	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"github.com/denwilliams/go-lifx-mqtt/internal/mqtt"
	"github.com/essentialkaos/ek/v12/color"
)

func NewClient() *LIFXClient {
	lights := make(lightMap)
	return &LIFXClient{lights: lights}
}

type LIFXClient struct {
	lights lightMap
}

func (lc *LIFXClient) TurnOn(id string) {
	lc.SetPowerState(id, true)
}

func (lc *LIFXClient) TurnOff(id string) {
	lc.SetPowerState(id, false)
}

func (lc *LIFXClient) SetWhite(id string, brightness uint16, kelvin uint16, duration uint32) {
	state := golifx.HSBK{
		Kelvin:     kelvin,
		Brightness: brightness,
	}
	lc.SetColorState(id, &state, duration)
}

func (lc *LIFXClient) SetColor(id string, hue uint16, saturation uint16, brightness uint16, duration uint32) {
	state := golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
	}
	lc.SetColorState(id, &state, duration)
}

func (lc *LIFXClient) SetColorState(id string, state *golifx.HSBK, duration uint32) {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("Bulb %s not found", id)
		return
	}
	if l.state != nil && l.state.Power {
		l.bulb.SetPowerDurationState(true, duration)
	}
	res, err := l.bulb.SetColorStateWithResponse(state, duration)
	if err == nil {
		l.state = res
	} else if err != io.EOF {
		logging.Error("Error setting color state: %v", err)
		l.bulb.SetPowerState(true)
		return
	}
	if res == nil || !res.Power {
		l.bulb.SetPowerDurationState(true, duration)
	}
}

func (lc *LIFXClient) HandleCommand(id string, command *mqtt.Command) error {
	if command == nil {
		return nil
	}

	duration := command.Duration
	if duration == 0 {
		duration = 1500
	}

	if command.Brightness == 0 {
		lc.TurnOff(id)
		return nil
	}

	if command.Temperature > 0 {
		// lc.TurnOn(id)

		b := uint16((float32(0xffff) * (float32(command.Brightness) / 100)))
		lc.SetWhite(id, b, command.Temperature, duration)
		logging.Info("Set bulb %s white %d%% %dK", id, command.Brightness, command.Temperature)
		return nil
	}

	if command.Color != "" {
		// lc.TurnOn(id)

		hex, _ := color.Parse(command.Color)
		hsl := hex.ToRGB().ToHSV()
		h := uint16(math.Mod(float64(0x10000)*float64(hsl.H)/360, float64(0x10000)))
		s := uint16((float32(0xffff) * (float32(hsl.S) / 100)))
		b := uint16((float32(0xffff) * (float32(hsl.V) / 100)))

		lc.SetColor(id, h, s, b, duration)
		logging.Info("Set bulb %s color %vK", id, hsl)
		return nil
	}

	return nil
}

func (lc *LIFXClient) SetPowerState(id string, state bool) {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("Bulb %s not found", id)
		return
	}
	l.bulb.SetPowerState(state)
	logging.Info("Turning %s %s", id, onOrOff(state))
}

// IsRoot returns true if the receiver is the address of the root module,
// or false otherwise.
func (lc *LIFXClient) Discover() {
	bulbs, err := golifx.LookupBulbs()
	if err != nil {
		logging.Error("Error getting LIFX lights: %s", err)
		return
	}

	logging.Info("Found LIFX lights: %d", len(bulbs))

	for _, bulb := range bulbs {
		id := strings.Replace(bulb.MacAddress(), ":", "", -1)
		ip := bulb.IP()
		existing := lc.lights.Has(id)
		lc.lights.Set(id, &light{bulb: bulb})
		if !existing {
			logging.Info("Found LIFX bulb: %s %s", id, ip)
		}
	}
}
