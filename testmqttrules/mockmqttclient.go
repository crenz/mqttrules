package testmqttrules

import (
	"github.com/eclipse/paho.mqtt.golang"
)

// Mock MQTT Client. Defined to avoid import cycle
type MockMqttClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) bool
	Unsubscribe(topics ...string) bool
}

type mockMqttClient struct {
	connected bool
}

func NewClient() MockMqttClient {
	c := &mockMqttClient{}
	c.connected = false

	return c
}

func (c *mockMqttClient) IsConnected() bool {
	return c.connected
}

func (c *mockMqttClient) Connect() bool {
	c.connected = true
	return true
}

func (c *mockMqttClient) Disconnect() {
	c.connected = false
}

func (c *mockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) bool {
	return true
}

func (c *mockMqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) bool {
	return true
}

func (c *mockMqttClient) Unsubscribe(topics ...string) bool {
	return true
}
