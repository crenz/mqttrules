package mqttrules

import (
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//import "github.com/robfig/cron"

// Interface for MQTT client used to interface with broker
type MqttClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) bool
	Unsubscribe(topics ...string) bool
}

// MQTT Rules Client
type Client interface {
	Connect() bool
	Subscribe() bool
	Listen()
	Disconnect()
	SetPrefix(prefix string)
	GetPrefix() string

	SetParameter(parameter string, value string)
	GetParameterValue(parameter string) string
	TriggerParameterUpdate(parameter string, value string)
	ReplaceParamsInString(in string) string

	AddParameterSubscription(topic string, parameter string)
	RemoveParameterSubscription(topic string, parameter string)
	AddRuleSubscription(topic string, ruleset string, rule string)
	RemoveRuleSubscription(topic string, ruleset string, rule string)

	ExecuteRule(ruleset string, rule string, triggerPayload string)
	Publish(topic string, qos byte, retained bool, payload string)
}

type rulesMap map[string]map[string]Rule

type subscriptions struct {
	//TODO Use slices instead of maps
	parameters map[string]bool
	rules      map[string]map[string]bool
}

type subscriptionsMap map[string]subscriptions

type client struct {
	mqttClient MqttClient
	messages   chan [2]string
	prefix     string

	parameters      parameterMap
	parameterValues map[string]interface{}
	rules           rulesMap
	subscriptions   subscriptionsMap

	regexParam     *regexp.Regexp
	regexRule      *regexp.Regexp
	messagehandler mqtt.MessageHandler
}

func (c *client) initialize() {
	c.SetPrefix("")
	c.parameters = make(parameterMap)
	c.parameterValues = make(map[string]interface{})
	c.rules = make(rulesMap)
	c.subscriptions = make(subscriptionsMap)
	c.messagehandler = func(client mqtt.Client, msg mqtt.Message) {
		c.messages <- [2]string{msg.Topic(), string(msg.Payload())}
	}
	c.messages = make(chan [2]string)
}

// Creates and initializes a new MQTT rules client
func NewClient(mqttClient MqttClient, prefix string) Client {
	c := &client{}
	c.initialize()
	c.mqttClient = mqttClient
	c.prefix = prefix

	return c
}

func (c *client) Connect() bool {
	log.Infoln("Connecting to MQTT broker")

	return c.mqttClient.Connect()
}

func (c *client) Subscribe() bool {
	return c.mqttClient.Subscribe("#", byte(1), c.messagehandler)
}

func (c *client) Publish(topic string, qos byte, retained bool, payload string) {

	if success := c.mqttClient.Publish(topic, 1, retained, payload); !success {
		log.Errorf("Error publishing MQTT topic [%s]", topic)
	}
}

func (c *client) Listen() {
	incoming := <-c.messages
	//log.Infof("Received [%s] %s\n", incoming[0], incoming[1])

	_, exists := c.subscriptions[incoming[0]]
	if exists {
		c.handleIncomingTrigger(incoming[0], incoming[1])
	}

	if res := c.regexParam.FindStringSubmatch(incoming[0]); res != nil {
		c.handleIncomingParam(res[1], incoming[1])
	}
	if res := c.regexRule.FindStringSubmatch(incoming[0]); res != nil {
		c.handleIncomingRule(res[1], res[2], incoming[1])
	}
}

func (c *client) handleIncomingTrigger(topic string, payload string) {
	for key := range c.subscriptions[topic].parameters {
		c.TriggerParameterUpdate(key, payload)
	}
	for key := range c.subscriptions[topic].rules {
		for subkey := range c.subscriptions[topic].rules[key] {
			c.ExecuteRule(key, subkey, payload)
		}
	}
}

func (c *client) ensureSubscription(topic string) bool {
	if c.mqttClient == nil {
		return false
	}

	if _, exists := c.subscriptions[topic]; !exists {
		c.subscriptions[topic] = subscriptions{parameters: make(map[string]bool), rules: make(map[string]map[string]bool)}
	}
	if len(c.subscriptions[topic].parameters) == 0 && len(c.subscriptions[topic].rules) == 0 {
		if success := c.mqttClient.Subscribe(topic, byte(1), c.messagehandler); !success {
			log.Errorln("Failed to add subscription [%s]")
			return false
		}
		log.Infof("Subscribed to MQTT topic [%s]", topic)
	}
	return true
}

func (c *client) contemplateUnsubscription(topic string) bool {
	if c.mqttClient == nil {
		return false
	}

	if len(c.subscriptions[topic].parameters) == 0 && len(c.subscriptions[topic].rules) == 0 {
		delete(c.subscriptions, topic)
		if success := c.mqttClient.Unsubscribe(topic); !success {
			log.Errorln("Failed to remove subscription [%s]")
			return false
		}
		log.Infof("Unsubscribed from MQTT topic [%s]", topic)
	}
	return true
}

func (c *client) handleIncomingParam(param string, value string) {
	c.SetParameter(param, value)
}

func (c *client) Disconnect() {
	log.Infoln("Disconnecting from MQTT broker")
	c.mqttClient.Disconnect()
}

func (c *client) SetPrefix(prefix string) {
	c.prefix = prefix

	c.regexParam = regexp.MustCompile(fmt.Sprintf("^%sparam/([^/]+)", c.prefix))
	c.regexRule = regexp.MustCompile(fmt.Sprintf("^%srule/([^/]+)/([^/]+)", c.prefix))
}

func (c *client) GetPrefix() string {
	return c.prefix
}
