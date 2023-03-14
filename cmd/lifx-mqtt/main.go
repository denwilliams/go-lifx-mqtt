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
	ld := lifx.NewClient()
	mc.Connect(ld)
	defer mc.Disconnect()

	mainLoop(ld)

	logging.Info("Terminating program")
}

func mainLoop(ld *lifx.LIFXClient) {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// It can take a few runs to discover all the lights
	for i := (5 * time.Second); i <= 80*time.Second; i = i * 2 {
		time.AfterFunc(i, ld.Discover)
	}

	// We want to continually call the Discover method at an interval
	// to pick up on new lights that come online
	tick := time.Tick(10 * time.Minute)

	for {
		select {
		case <-tick:
			ld.Discover()
		case <-signalChan:
			// Stop the loop when an interrupt signal is received
			return
		}
	}
}
