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
	return true
}

func (c *pahoClient) Disconnect() {
	c.c.Disconnect(250)
}

func (c *pahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) bool {
	if token := c.c.Publish(topic, qos, retained, payload); token.Wait() && token.Error() != nil {
		log.Errorf("[PahoClient] Error publishing message for topic [%s]: %v", topic, token.Error())
		return false
	}
	return true
}

func (c *pahoClient) Subscribe(topic string, qos byte, callback func(string, string)) bool {
	pahoCallback := func(c mqtt.Client, m mqtt.Message) {
		callback(m.Topic(), string(m.Payload()))
	}

	if token := c.c.Subscribe(topic, qos, pahoCallback); token.Wait() && token.Error() != nil {
		log.Errorf("[PahoClient] Error subscribing to topic [%s]: %v", topic, token.Error())
		return false
	}
	log.Infof("[PahoClient] Subscribed to MQTT topic [%s]", topic)
	return true
}

func (c *pahoClient) Unsubscribe(topics ...string) bool {
	if token := c.c.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		log.Errorf("[PahoClient] Error unsubscribing from topic [%s]: %v", token, token.Error())
		return false
	}
	return true
}
