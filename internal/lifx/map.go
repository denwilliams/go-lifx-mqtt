package lifx

import (
	"github.com/2tvenom/golifx"
)

type light struct {
	bulb  *golifx.Bulb
	state *golifx.BulbState
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
