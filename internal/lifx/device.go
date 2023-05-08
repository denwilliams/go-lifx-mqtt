package lifx

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	lifxlight "github.com/denwilliams/go-lifx-mqtt/internal/lifx/light"
	lifxrelay "github.com/denwilliams/go-lifx-mqtt/internal/lifx/relay"
	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"go.yhsif.com/lifxlan"
)

func newDevice(id string, device lifxlan.Device) *lifxdevice {
	return &lifxdevice{id: id, device: device}
}

type lifxdevice struct {
	id         string
	loaded     bool
	device     lifxlan.Device
	light      lifxlight.Device
	relay      lifxrelay.Device
	product    *lifxlan.Product
	power      lifxlan.Power
	color      *lifxlan.Color
	relayPower [4]lifxlan.Power
	mu         sync.Mutex
	timer      *time.Timer
}

func (l *lifxdevice) Load() error {
	if l.device == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	logging.Debug("Loading %s", l.id)

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	d := l.device

	conn, err := d.Dial()
	if err != nil {
		// TODO: This doesn't need to be fatal
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}
	if err := d.GetHardwareVersion(ctx, conn); err != nil {
		logging.Warn("Failed to get hardware version %s", l.id)
		return err
	}

	logging.Debug("Loaded %s type=%d", l.id, (d.HardwareVersion().ProductID))

	lifxType, product := getType(d.HardwareVersion())
	l.product = product

	if lifxType == Light {
		logging.Debug("Wrapping %s light", l.id)

		l.light = lifxlight.Wrap(l.device)
		l.loaded = true
		return nil
	}

	if lifxType == Switch {
		logging.Debug("Wrapping %s relay", l.id)

		l.relay = lifxrelay.Wrap(l.device)
		l.loaded = true
		return nil
	}

	logging.Warn("Ignoring wrapping device %s type=%d", l.id, lifxType)
	l.loaded = true
	return nil
}

func (l *lifxdevice) Refresh(emitter StatusEmitter) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	logging.Info("Refreshing %s", l.id)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := l.device.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		// shouldn't be fatal
		log.Fatal(ctx.Err())
	}

	power, errP := l.device.GetPower(ctx, conn)
	if errP != nil {
		logging.Warn("Failed to get power %s %s", l.id, errP.Error())
		return errP
	}
	if l.power != power {
		l.power = power
		logging.Debug("Refreshed %s power=%v", l.id, power)
		emitter.EmitStatus(ctx, l.id, "power", toPowerPayload(power))
	}

	if l.light != nil {
		color, errC := l.light.GetColor(ctx, conn)
		if errC != nil {
			logging.Warn("Failed to get color %s %s", l.id, errC.Error())
			return errC
		}
		if !isSameColor(l.color, color) {
			l.color = color
			logging.Debug("Refreshed %s color=%v", l.id, *color)
			emitter.EmitStatus(ctx, l.id, "color", toColorPayload(color))
		}
	}

	if l.relay != nil {
		for i := uint8(0); i < 4; i++ {
			power, errR := l.relay.GetRPower(ctx, conn, i)
			if errR != nil {
				logging.Warn("Failed to get relay %s %s", l.id, errR.Error())
			}
			if l.relayPower[i] != power {
				l.relayPower[i] = power
				emitter.EmitStatus(ctx, l.id, "relay"+strconv.Itoa(int(i)), toPowerPayload(power))
			}
		}
		logging.Debug("Refreshed %s relayPower=%v", l.id, l.relayPower)
	}

	return nil
}

func (l *lifxdevice) QueueRefresh(emitter StatusEmitter, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.timer != nil {
		l.timer.Stop()
	}
	if duration == 0 {
		duration = 1 * time.Second
	}
	l.timer = time.AfterFunc(duration, func() {
		l.Refresh(emitter)
	})
}
