package lifx

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"github.com/denwilliams/go-lifx-mqtt/internal/mqtt"
	"github.com/icza/gox/imagex/colorx"
	"go.yhsif.com/lifxlan"
)

var (
	defaultDuration uint32 = 1500
)

func NewClient(emitter StatusEmitter) *LIFXClient {
	lights := make(deviceMap)
	return &LIFXClient{devices: lights, emitter: emitter}
}

type LIFXClient struct {
	devices     deviceMap
	discovering bool
	emitter     StatusEmitter
}

func (lc *LIFXClient) AddDevice(ip string, mac string) error {
	key := strings.Replace(mac, ":", "", -1)
	t, err := lifxlan.ParseTarget(mac)
	if err != nil {
		return err
	}
	logging.Debug("Adding device %s %s %s", key, ip, t)
	addr := net.JoinHostPort(ip, lifxlan.DefaultBroadcastPort)
	d := lifxlan.NewDevice(addr, lifxlan.ServiceUDP, t)

	l := newDevice(key, d)
	lc.devices.Set(key, l)
	return l.Load()
}

func (lc *LIFXClient) Discover() {
	timeout := 30 * time.Second
	lc.DiscoverWithTimeout(timeout)
}

func (lc *LIFXClient) DiscoverWithTimeout(timeout time.Duration) int {
	numDiscovered := 0

	if lc.discovering {
		logging.Warn("Aborted - already discovering")
		return 0
	}

	lc.discovering = true

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	deviceChan := make(chan lifxlan.Device)

	go func() {
		if err := lifxlan.Discover(ctx, deviceChan, ""); err != nil {
			if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
				log.Fatalf("Discover failed: %v", err)
			}
		}
	}()

	for device := range deviceChan {
		t := device.Target().String()
		key := strings.Replace(t, ":", "", -1)

		if lc.devices.Has(key) {
			continue
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		if err := device.GetLabel(ctx, nil); err != nil {
			logging.Warn("Couldn't get label for device=%s err=%s", t, err.Error())
			continue
		}

		l := newDevice(key, device)
		lc.devices.Set(key, l)
		numDiscovered++
		logging.Info("Found device label=\"%s\" target=%s", device.Label(), t)
	}

	logging.Debug("Total lights discovered: %d", len(lc.devices))
	lc.discovering = false

	return numDiscovered
}

func (lc *LIFXClient) LoadDevices() {
	for _, l := range lc.devices {
		if !l.loaded {
			go l.Load()
		}
	}
}

func (lc *LIFXClient) RefreshDevices() {
	for _, l := range lc.devices {
		l.QueueRefresh(lc.emitter, 0)
	}
}

func (lc *LIFXClient) TurnOn(id string, duration uint32) error {
	l := lc.devices.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	time := time.Duration(duration) * time.Millisecond

	defer l.QueueRefresh(lc.emitter, time)

	if l.light != nil {
		return l.light.SetLightPower(ctx, nil, lifxlan.PowerOn, time, true)
	}

	return l.device.SetPower(ctx, nil, lifxlan.PowerOn, true)
}

func (lc *LIFXClient) TurnOff(id string, duration uint32) error {
	l := lc.devices.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	time := time.Duration(duration) * time.Millisecond

	defer l.QueueRefresh(lc.emitter, time)

	if l.light != nil {
		return l.light.SetLightPower(ctx, nil, lifxlan.PowerOff, time, true)
	}

	return l.device.SetPower(ctx, nil, lifxlan.PowerOff, true)
}

func (lc *LIFXClient) SetWhite(id string, brightness uint16, kelvin uint16, duration uint32) error {
	l := lc.devices.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if l.light == nil {
		return nil
	}

	conn, err := l.light.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	b := uint16((float32(0xffff) * (float32(brightness) / 100)))
	if brightness == 0 && l.color != nil {
		b = l.color.Brightness
	}

	if kelvin == 0 && l.color != nil {
		kelvin = l.color.Kelvin
	}

	time := time.Duration(duration) * time.Millisecond

	hsbk := &lifxlan.Color{
		Kelvin:     kelvin,
		Brightness: b,
	}

	defer l.QueueRefresh(lc.emitter, time)

	err = l.light.SetColor(ctx, conn, hsbk, time, true)
	if err != nil {
		return err
	}

	err = l.light.SetPower(ctx, conn, lifxlan.PowerOn, true)
	if err != nil {
		return err
	}

	return nil
}

func (lc *LIFXClient) SetColor(id string, hsbk *lifxlan.Color, duration uint32) error {
	l := lc.devices.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if l.light == nil {
		return nil
	}

	// Mutating input - yuk - best find another way
	if hsbk.Kelvin == 0 && l.color != nil {
		hsbk.Kelvin = l.color.Kelvin
	}

	conn, err := l.light.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	time := time.Duration(duration) * time.Millisecond

	defer l.QueueRefresh(lc.emitter, time)

	err = l.light.SetColor(ctx, conn, hsbk, time, true)
	if err != nil {
		return err
	}

	err = l.light.SetPower(ctx, conn, lifxlan.PowerOn, true)
	if err != nil {
		return err
	}

	return nil
}

func (lc *LIFXClient) SetRelay(id string, index uint8, power bool) error {
	l := lc.devices.Get(id)
	if l == nil {
		logging.Warn("No device found for id=%s", id)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if l.relay == nil {
		return nil
	}

	defer l.QueueRefresh(lc.emitter, 100*time.Millisecond)

	if err := l.relay.SetRPower(ctx, nil, index, getPower(power), true); err != nil {
		return err
	}

	return nil
}

func (lc *LIFXClient) HandleCommand(id string, command *mqtt.Command) error {
	if id == "discover" {
		go lc.Discover()
		return nil
	}

	if command == nil {
		return nil
	}

	dur := defaultDuration
	if command.Duration != nil {
		dur = *command.Duration
	}

	brightness := uint16(0)
	if command.Brightness != nil {
		brightness = *command.Brightness
		if brightness == 0 {
			return lc.TurnOff(id, dur)
		}
	}
	temperature := uint16(0)
	if command.Temperature != nil {
		temperature = *command.Temperature
	}

	if temperature > 0 {
		logging.Info("Set light %s %dK %d%%", id, temperature, brightness)
		lc.SetWhite(id, brightness, temperature, dur)
		return nil
	} else if brightness > 0 {
		logging.Info("Set light %s %dK %d%%", id, temperature, brightness)
		lc.SetWhite(id, brightness, temperature, dur)
		return nil
	}

	if command.Color != nil {
		color := *command.Color

		c, err := colorx.ParseHexColor(color)
		if err != nil {
			logging.Warn("Error parsing color %s err=%s", command.Color, err)
			return nil
		}

		hsbk := lifxlan.FromColor(c, temperature)
		logging.Info("Set light %s color %v", id, *hsbk)
		return lc.SetColor(id, hsbk, dur)
	}

	if command.Relay0 != nil {
		logging.Info("Set relay0 %s %v", id, *command.Relay0)
		lc.SetRelay(id, 0, *command.Relay0)
	}
	if command.Relay1 != nil {
		logging.Info("Set relay1 %s %v", id, *command.Relay1)
		lc.SetRelay(id, 1, *command.Relay1)
	}
	if command.Relay2 != nil {
		logging.Info("Set relay2 %s %v", id, *command.Relay2)
		lc.SetRelay(id, 2, *command.Relay2)
	}
	if command.Relay3 != nil {
		logging.Info("Set relay3 %s %v", id, *command.Relay3)
		lc.SetRelay(id, 3, *command.Relay3)
	}

	return nil
}

func getPower(power bool) lifxlan.Power {
	if power {
		return lifxlan.PowerOn
	}
	return lifxlan.PowerOff
}
