package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/crenz/mqttrules/agent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mattn/go-colorable"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(colorable.NewColorableStdout())
}

func main() {
	pBroker := flag.String("broker", "tcp://localhost:1883", "MQTT broker URI (e.g. tcp://localhost:1883)")
	pPassword := flag.String("config", "", "(optional) configuration file")
	pUsername := flag.String("username", "", "(optional) user name for MQTT broker access")

	flag.Parse()

	log.Infoln("mqtt-rules connecting to broker", *pBroker)
	opts := mqtt.NewClientOptions()
	opts.SetClientID("mqtt-rules/0.1")
	opts.AddBroker(*pBroker)
	opts.SetUsername(*pUsername)
	opts.SetPassword(*pPassword)
	mqttClient := NewPahoClient(opts)

	a := agent.New(mqttClient, "")

	if a.Connect() {
		a.Subscribe()

		for {
			a.Listen()
		}
		// will never be reached: c.Disconnect()
	}
}

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
	return true
}

func (c *pahoClient) Unsubscribe(topics ...string) bool {
	if token := c.c.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		log.Errorf("[PahoClient] Error unsubscribing from topic [%s]: %v", token, token.Error())
		return false
	}
	return true
}
