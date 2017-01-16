package agent

import (
	log "github.com/Sirupsen/logrus"
	"github.com/eclipse/paho.mqtt.golang"
)

type PahoClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte, callback func(string, string)) bool
	Unsubscribe(topics ...string) bool
}

type pahoClient struct {
	c mqtt.Client
}

func NewPahoClient(o *mqtt.ClientOptions) PahoClient {
	c := &pahoClient{}
	c.c = mqtt.NewClient(o)
	return c
}

func (c *pahoClient) IsConnected() bool {
	return c.c.IsConnected()
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

func (c *pahoClient) Subscribe(topic string, qos byte, callback func(string, string)) bool {
	pahoCallback := func(c mqtt.Client, m mqtt.Message) {
		callback(m.Topic(), string(m.Payload()))
	}

	if token := c.c.Subscribe(topic, qos, pahoCallback); token.Wait() && token.Error() != nil {
		log.WithFields(log.Fields{
			"component": "PahoClient",
			"topic":     topic,
			"error":     token.Error(),
		}).Error("Error subscribing to MQTT topic")
		return false
	}
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
	log.WithFields(log.Fields{
		"component": "PahoClient",
		"topic":     topics,
	}).Debug("Unsubscribed from MQTT topics")
	return true
}
