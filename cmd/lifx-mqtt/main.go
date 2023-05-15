package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/denwilliams/go-lifx-mqtt/internal/lifx"
	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	"github.com/denwilliams/go-lifx-mqtt/internal/mqtt"
	"github.com/denwilliams/go-lifx-mqtt/internal/web"
	"github.com/joho/godotenv"
)

func init() {
	logging.Init(nil, logging.DefaultFlags)
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
	baseTopic := os.Getenv("MQTT_TOPIC_PREFIX")
	subscribeTopic := os.ExpandEnv("$MQTT_TOPIC_PREFIX/set/#")
	portStr := os.Getenv("PORT")
	serverPort, _ := strconv.Atoi(portStr)
	if err != nil {
		logging.Error("Error parsing HTTP_PORT %s", err)
	}

	mc := mqtt.NewMQTTClient(mu, baseTopic, subscribeTopic)
	lc := lifx.NewClient(mqtt.NewMqttStatusEmitter(mc))
	mc.Connect(lc)
	defer mc.Disconnect()

	go loadDevices(lc)
	go updateCache(lc)
	go discoverLoop(lc)
	// NOTE: can use AddDevice to avoid having to rediscover each startup
	// err = lc.AddDevice("1.2.3.4:1234", "0:73:d5:01:23:45")
	if serverPort > 0 {
		go startServer(serverPort)
	}

	logging.Info("Ready")

	waitForExit()

	logging.Info("Terminating")
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

func loadDevices(lc *lifx.LIFXClient) {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	tick := time.Tick(15 * time.Second)

	for {
		select {
		case <-tick:
			lc.LoadDevices()
		case <-signalChan:
			// Stop the loop when an interrupt signal is received
			logging.Info("Background device loader interrupted, exiting")
			return
		}
	}
}

func updateCache(lc *lifx.LIFXClient) {
	// Set up a channel to receive OS signals so we can gracefully exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	tick := time.Tick(10 * time.Minute)

	for {
		select {
		case <-tick:
			lc.RefreshDevices()
		case <-signalChan:
			// Stop the loop when an interrupt signal is received
			logging.Info("Background cached state updater interrupted, exiting")
			return
		}
	}
}

func startServer(port int) {
	logging.Info("Creating HTTP server")
	handler := web.CreateHandler()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}
	logging.Info("Starting HTTP server on port %d", port)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
		log.Fatal(err)
	}
}
