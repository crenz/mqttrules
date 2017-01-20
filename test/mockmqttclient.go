package test

type MockMessageHandler func(topic string, payload string)

// MockMqttClient provides a mock client for testing. Defined to avoid import cycle.
type MockMqttClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	SetSubscriptionCallback(callback func(string, string))
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte) bool
	Unsubscribe(topics ...string) bool
	IsSubscribed(topic string) bool
	LastMessage() MockMqttMessage
}

// MockMqttMessage provides a mock MQTT message structure for testing. Defined to avoid import cycle.
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
	callback      MockMessageHandler
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

func (c *mockMqttClient) SetSubscriptionCallback(callback func(string, string)) {
	c.callback = callback
}

func (c *mockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) bool {
	c.lastMessage = MockMqttMessage{
		Topic:    topic,
		QoS:      qos,
		Retained: retained,
		Payload:  payload,
	}

	if c.callback != nil {
		go c.callback(topic, payload.(string))
	}
	return true
}

func (c *mockMqttClient) Subscribe(topic string, qos byte) bool {
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
