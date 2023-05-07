package lifx

import (
	lifxlanlight "go.yhsif.com/lifxlan/light"
)

type light struct {
	device *lifxlanlight.Device
}

type lightMap map[string]*light

func (lm *lightMap) Get(id string) *light {
	b, ok := (*lm)[id]
	if !ok {
		return nil
	}

	return b
}

func (lm *lightMap) Set(id string, l *light) {
	(*lm)[id] = l
}

func (lm *lightMap) Delete(id string) {
	delete(*lm, id)
}

func (lm *lightMap) Has(id string) bool {
	_, exists := (*lm)[id]
	return exists
}
