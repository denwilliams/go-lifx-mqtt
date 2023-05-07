package main

import (
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denwilliams/go-lifx-mqtt/internal/lifx"
	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"github.com/denwilliams/go-lifx-mqtt/internal/mqtt"
	"github.com/joho/godotenv"
)

func init() {
	logging.Init()
	logging.Info("Loading .env file")
	err := godotenv.Load(".env")

	if err != nil {
		logging.Warn("Unable to load .env")
	}
}

func main() {
	mu, err := url.Parse(os.Getenv("MQTT_URI"))
	if err != nil {
		logging.Error("Error parsing URL %s", err)
	}
	topic := os.ExpandEnv("$MQTT_TOPIC_PREFIX/set/#")

	mc := mqtt.NewMQTTClient(mu, topic)
	lc := lifx.NewClient()
	mc.Connect(lc)
	defer mc.Disconnect()

	go discoverLoop(lc)
	go updateCachedState(lc)
	// NOTE: can use NewDevice to avoid having to rediscover each startup
	// eg NewDevice("1.2.3.4:1234", lifxlan.ServiceUDP, ParseTarget("0123456"))

	logging.Info("Ready")

	waitForExit()

	logging.Info("Terminating program")
}

func waitForExit() {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	logging.Info("Exit signal received")
}

func discoverLoop(lc *lifx.LIFXClient) {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	logging.Info("Performing initial discovery")
	// It can take a few runs to discover all the lights
	// Keep going until we find no new lights for a few runs
	emptyRuns := 0
	for emptyRuns < 10 {
		found := lc.DiscoverWithTimeout(15 * time.Second)
		if found == 0 {
			emptyRuns++
		} else {
			emptyRuns = 0
		}
	}
	logging.Info("Finished initial light discovery, will continue to discover every 10 minutes")

	// We want to continually call the Discover method at an interval
	// to pick up on new lights that come online
	tick := time.Tick(10 * time.Minute)

	for {
		select {
		case <-tick:
			lc.DiscoverWithTimeout(60 * time.Second)
		case <-signalChan:
			// Stop the loop when an interrupt signal is received
			logging.Info("Background discovery loop interrupted, exiting")
			return
		}
	}
}

func updateCachedState(lc *lifx.LIFXClient) {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	tick := time.Tick(1 * time.Minute)

	for {
		select {
		case <-tick:
			lc.RefreshLightState()
		case <-signalChan:
			// Stop the loop when an interrupt signal is received
			logging.Info("Background cached state updater interrupted, exiting")
			return
		}
	}
}
