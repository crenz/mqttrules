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
	IsSubscribed(topic string) bool
	LastMessage() MockMqttMessage
}

type MockMqttMessage struct {
	Topic    string
	QoS      byte
	Retained bool
	Payload  interface{}
}

type mockMqttClient struct {
	connected     bool
	subscriptions map[string]bool
	lastMessage   MockMqttMessage
}

func NewClient() MockMqttClient {
	c := &mockMqttClient{}
	c.subscriptions = make(map[string]bool)
	c.connected = false
	c.lastMessage = MockMqttMessage{}

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
	c.lastMessage = MockMqttMessage{
		Topic:    topic,
		QoS:      qos,
		Retained: retained,
		Payload:  payload,
	}
	return true
}

func (c *mockMqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) bool {
	c.subscriptions[topic] = true
	return true
}

func (c *mockMqttClient) Unsubscribe(topics ...string) bool {
	for _, topic := range topics {
		delete(c.subscriptions, topic)
	}
	return true
}

func (c *mockMqttClient) IsSubscribed(topic string) bool {
	return c.subscriptions[topic]
}

func (c *mockMqttClient) LastMessage() MockMqttMessage {
	return c.lastMessage
}
