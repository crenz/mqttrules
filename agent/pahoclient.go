package agent

import (
	log "github.com/Sirupsen/logrus"
	"github.com/eclipse/paho.mqtt.golang"
)

type PahoClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	SetSubscriptionCallback(callback func(string, string))
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte) bool
	Unsubscribe(topics ...string) bool
}

type pahoClient struct {
	c                    mqtt.Client
	subscriptionCallback func(string, string)
	subscriptions        map[string]byte
}

func NewPahoClient(o *mqtt.ClientOptions) PahoClient {
	c := &pahoClient{}
	c.subscriptions = make(map[string]byte)

	pahoOnConnectCallback := func(cm mqtt.Client) {
		log.WithFields(log.Fields{
			"component": "PahoClient",
		}).Debug("Established connection to broker")
		c.resubscribe()
	}
	o.SetOnConnectHandler(pahoOnConnectCallback)

	c.c = mqtt.NewClient(o)
	return c
}

func (c *pahoClient) IsConnected() bool {
	return c.c.IsConnected()
}

func (c *pahoClient) resubscribe() {
	for topic, qos := range c.subscriptions {
		c.Subscribe(topic, qos)
	}
}

func (c *pahoClient) Connect() bool {
	if token := c.c.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("[PahoClient] Error connecting to MQTT broker: %v", token.Error())
		return false
	}
	log.WithFields(log.Fields{
		"component": "PahoClient",
	}).Debug("Connected to MQTT broker successfully")
	return true
}

func (c *pahoClient) Disconnect() {
	c.c.Disconnect(250)
}

func (c *pahoClient) SetSubscriptionCallback(callback func(string, string)) {
	c.subscriptionCallback = callback
}

func (c *pahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) bool {
	if token := c.c.Publish(topic, qos, retained, payload); token.Wait() && token.Error() != nil {
		log.WithFields(log.Fields{
			"component": "PahoClient",
			"topic":     topic,
			"error":     token.Error(),
		}).Error("Error publishing message")
		return false
	}
	log.WithFields(log.Fields{
		"component": "PahoClient",
		"topic":     topic,
	}).Debug("Published message")
	return true
}

func (c *pahoClient) Subscribe(topic string, qos byte) bool {
	pahoCallback := func(cm mqtt.Client, m mqtt.Message) {
		c.subscriptionCallback(m.Topic(), string(m.Payload()))
	}

	if token := c.c.Subscribe(topic, qos, pahoCallback); token.Wait() && token.Error() != nil {
		log.WithFields(log.Fields{
			"component": "PahoClient",
			"topic":     topic,
			"error":     token.Error(),
		}).Error("Error subscribing to MQTT topic")
		return false
	}
	c.subscriptions[topic] = qos
	log.WithFields(log.Fields{
		"component": "PahoClient",
		"topic":     topic,
	}).Debug("Subscribed to MQTT topic")
	return true
}

func (c *pahoClient) Unsubscribe(topics ...string) bool {
	if token := c.c.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		log.WithFields(log.Fields{
			"component": "PahoClient",
			"topic":     topics,
			"error":     token.Error(),
		}).Error("Error unsubscribingfrom MQTT topics")
		return false
	}
	for _, topic := range topics {
		delete(c.subscriptions, topic)
	}
	log.WithFields(log.Fields{
		"component": "PahoClient",
		"topic":     topics,
	}).Debug("Unsubscribed from MQTT topics")
	return true
}
