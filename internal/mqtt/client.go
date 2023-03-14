package mqtt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/denwilliams/go-lifx-mqtt/internal/logging"
	pm "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client *pm.Client
	topic  string
}

func (mc *MQTTClient) Publish(payload string) {
	// Publish a message to the topic with a QoS of 1
	if token := (*mc.client).Publish(mc.topic, 1, false, payload); token.Wait() && token.Error() != nil {
		// TODO: don't panic, just return
		panic(token.Error())
	}
}

func (mc *MQTTClient) Connect(h CommandHandler) {
	// Connect to the MQTT broker
	if token := (*mc.client).Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	prefix := strings.Replace(mc.topic, "#", "", 1)

	// Set up a callback function to handle incoming messages
	messageHandler := func(client pm.Client, msg pm.Message) {
		topic := msg.Topic()
		if !strings.HasPrefix(topic, prefix) {
			return
		}
		id := strings.Replace(topic, prefix, "", 1)

		bytes := msg.Payload()
		payload, err := parsePayload(&bytes)
		if err != nil {
			fmt.Printf("Error unmarshalling JSON: %s %v\n", err, string(bytes))
			return
		}
		fmt.Printf("Received message on topic %s: %s\n", id, payload.String())

		// messages := make(chan string)
		// go func() { messages <- "ping" }()
		// msg := <-messages

		go func() {
			h.HandleCommand(id, payload)
		}()
	}

	// Subscribe to the topic with a QoS of 1
	if token := (*mc.client).Subscribe(mc.topic, 1, messageHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (mc *MQTTClient) Disconnect() {
	logging.Info("Disconnecting from MQTT")

	// Unsubscribe from the topic
	if token := (*mc.client).Unsubscribe(mc.topic); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Disconnect from the MQTT broker
	(*mc.client).Disconnect(250)
}

func NewMQTTClient(uri *url.URL, topic string) *MQTTClient {
	// Create a new MQTT client with the default options
	opts := pm.NewClientOptions().AddBroker(uri.String()).SetClientID("emqx_test_client")
	client := pm.NewClient(opts)
	return &MQTTClient{client: &client, topic: topic}
}

func parsePayload(bytes *[]byte) (*Command, error) {
	var payload Command
	if err := json.Unmarshal(*bytes, &payload); err == nil {
		return &payload, nil
	}

	var passOne string
	if err := json.Unmarshal(*bytes, &passOne); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(passOne), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}