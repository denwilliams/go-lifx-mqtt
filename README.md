# go-lifx-mqtt

**NOTE: This is a work in progress.**

This is a simple MQTT client that listens for messages containing commands and sends them to the LIFX bulbs.

Initially designed to be a replacement for https://github.com/denwilliams/lifx-mqtt. Future versions may support more features.

## Running

run `make run` or `go run ./cmd/lifx-mqtt`

## Building

run `make build` or `go build -o lifx-mqtt ./cmd/lifx-mqtt`

## Configuration

Using environment variables:

```bash
MQTT_URI="mqtt://localhost:1883" MQTT_TOPIC_PREFIX="lifx" ./lifx-mqtt
```

or from source..

```bash
MQTT_URI="mqtt://localhost:1883" MQTT_TOPIC_PREFIX="lifx" go run ./cmd/lifx-mqtt
```

For development you can create a `.env` file in the root of the project.

## Topics

Assuming the `MQTT_TOPIC_PREFIX` is `lifx`:

### `lifx/set/discover`

Starts a discovery of LIFX bulbs on the network. This is done automatically on startup, and periodically in the background, but can be triggered manually if needed.


### `lifx/set/{id}`

Set the state of a bulb matching {id}, where {id} can be seen in the app and is derived from the MAC address (all lower case, without `:` characters).

#### Payload Examples

Light Off:

```json
{
  "brightness": 0
}
```

Light Warm Full Brightness:

```json
{
  "brightness": 100,
  "temp": 2700
}
```

Light Cool Full Brightness:

```json
{
  "brightness": 100,
  "temp": 6500
}
```

Light Red:

```json
{"color": "#FF0000"}
```

Light Green:

```json
{"color": "#00FF00"}
```

Fade the light out over 10s:

```json
{
  "brightness": 0,
  "duration": 10000
}
```

### TODO `lifx/set/{id}/(on|off)`

TODO: turn a light on/off without messing with it's state
