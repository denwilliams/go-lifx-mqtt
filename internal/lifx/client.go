package lifx

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"github.com/denwilliams/go-lifx-mqtt/internal/mqtt"
	"github.com/icza/gox/imagex/colorx"
	"go.yhsif.com/lifxlan"
	lifxlanlight "go.yhsif.com/lifxlan/light"
)

var (
	defaultDuration uint32 = 1500
)

func NewClient() *LIFXClient {
	lights := make(lightMap)
	return &LIFXClient{lights: lights}
}

type LIFXClient struct {
	lights      lightMap
	discovering bool
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

		if lc.lights.Has(key) {
			continue
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		if err := device.GetLabel(ctx, nil); err != nil {
			logging.Warn("Couldn't get label for device=%s err=%s", t, err.Error())
			continue
		}

		l, err := lifxlanlight.Wrap(ctx, device, false)
		if checkContextError(err) {
			log.Printf("Check light capabilities for %v failed: %v", device, err)
			cancel()
			continue
		} else if l == nil {
			logging.Warn("Cast to light failed for %s", t)
			continue
		}

		numDiscovered++
		logging.Info("Found device label=\"%s\" target=%s", device.Label(), t)
		lc.lights.Set(key, &light{device: &l})
	}

	logging.Debug("Done - total lights discovered: %d", len(lc.lights))
	lc.discovering = false
	return numDiscovered
}

func (lc *LIFXClient) TurnOn(id string) error {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	d := *l.device
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return d.SetPower(ctx, nil, lifxlan.PowerOn, true)
}

func (lc *LIFXClient) TurnOff(id string) error {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	d := *l.device
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return d.SetPower(ctx, nil, lifxlan.PowerOff, true)
}

func (lc *LIFXClient) SetWhite(id string, brightness uint16, kelvin uint16, duration uint32) error {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	d := *l.device
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	b := uint16((float32(0xffff) * (float32(brightness) / 100)))
	time := time.Duration(duration) * time.Millisecond

	hsbk := &lifxlan.Color{
		Kelvin:     kelvin,
		Brightness: b,
	}

	logging.Info("Duration %d", duration)
	err = d.SetColor(ctx, conn, hsbk, time, true)
	if err != nil {
		return err
	}

	err = d.SetPower(ctx, conn, lifxlan.PowerOn, true)
	if err != nil {
		return err
	}

	return nil
}

func (lc *LIFXClient) SetColor(id string, hsbk *lifxlan.Color, duration uint32) error {
	l := lc.lights.Get(id)
	if l == nil {
		logging.Warn("No light found for id=%s", id)
		return nil
	}

	d := *l.device
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	time := time.Duration(duration) * time.Millisecond

	err = d.SetColor(ctx, conn, hsbk, time, true)
	if err != nil {
		return err
	}

	err = d.SetPower(ctx, conn, lifxlan.PowerOn, true)
	if err != nil {
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

	dur := command.Duration
	if dur == 0 {
		dur = defaultDuration
	}

	brightness := uint16(0)
	if command.Brightness != nil {
		brightness = *command.Brightness
		if brightness == 0 {
			return lc.TurnOff(id)
		}
	}

	temperature := uint16(0)
	if command.Temperature != nil {
		temperature = *command.Temperature
	}
	if temperature > 0 {
		logging.Info("Set bulb %s %dK %d%%", id, temperature, brightness)
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
		logging.Info("Set bulb %s color %v", id, *hsbk)
		return lc.SetColor(id, hsbk, dur)
	}

	return nil
}

func checkContextError(err error) bool {
	return err != nil && err != context.Canceled && err != context.DeadlineExceeded
}

func safeUint16(s *uint16) string {
	if s == nil {
		return "(nil)"
	}
	return fmt.Sprintf("%d", *s)
}

func safeString(s *string) string {
	if s == nil {
		return "(nil)"
	}
	return (*s)
}
