package lifx

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	devicesControlled = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lifx_devices_controlled_total",
		Help: "The total number of LIFX devices controlled",
	}, []string{"device_type", "on_state"})
)
