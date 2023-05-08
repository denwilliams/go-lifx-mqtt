package lifx

type deviceMap map[string]*lifxdevice

func (lm *deviceMap) Get(id string) *lifxdevice {
	b, ok := (*lm)[id]
	if !ok {
		return nil
	}

	return b
}

func (lm *deviceMap) Set(id string, l *lifxdevice) {
	(*lm)[id] = l
}

func (lm *deviceMap) Delete(id string) {
	delete(*lm, id)
}

func (lm *deviceMap) Has(id string) bool {
	_, exists := (*lm)[id]
	return exists
}
