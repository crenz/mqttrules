package mqttrules

import (
	"encoding/json"
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//import "github.com/robfig/cron"

type Client interface {
	Connect(broker string, username string, password string) bool
	Subscribe() bool
	Listen()
	Disconnect()
	SetPrefix(prefix string)
	GetPrefix() string

	SetParameter(parameter string, value string)
	GetParameterValue(parameter string) string
	TriggerParameterUpdate(parameter string, value string)
	ReplaceParamsInString(in string) string

	SetRule(id string, conditions []string, actions []Action)
	GetRule(id string) Rule
	ExecuteRule(id string) Action

	AddParameterSubscription(topic string, parameter string)
	RemoveParameterSubscription(topic string, parameter string)
}

type rulesMap map[string]map[string]Rule

type subscriptions struct {
	//TODO Use slices instead of maps
	parameters map[string]bool
	rules      map[string]map[string]bool
}

type subscriptionsMap map[string]subscriptions

type client struct {
	mqttClient mqtt.Client
	messages   chan [2]string
	prefix     string

	parameters    parameterMap
	rules         rulesMap
	subscriptions subscriptionsMap

	regexParam *regexp.Regexp
	regexRule  *regexp.Regexp
}

func (c *client) initialize() {
	c.SetPrefix("")
	c.parameters = make(parameterMap)
	c.rules = make(rulesMap)
	c.subscriptions = make(subscriptionsMap)
}

func NewClient() Client {
	c := &client{}
	c.initialize()

	return c
}

func (c *client) Connect(broker string, username string, password string) bool {
	log.Infoln("Connecting to MQTT broker", broker)

	opts := mqtt.NewClientOptions()
	opts.SetClientID("mqtt-rules/0.1")
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)

	c.messages = make(chan [2]string)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		c.messages <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	c.mqttClient = mqtt.NewClient(opts)
	token := c.mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		log.Errorln(token.Error())
		return false
	}

	return true
}

func (c *client) Subscribe() bool {
	if token := c.mqttClient.Subscribe("#", byte(1), nil); token.Wait() && token.Error() != nil {
		log.Errorln(token.Error())
		return false
	}
	log.Infoln("Subscribed successfully")
	return true
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
		fmt.Println("TODO: Execute rule", key)
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
		if token := c.mqttClient.Subscribe(topic, byte(1), nil); token.Wait() && token.Error() != nil {
			log.Errorln("Failed to add subscription [%s]: %v", topic, token.Error())
			return false
		}
		log.Infof("Subscribed to MQTT topic [%s]", topic)
	}
	return true
}

func (c *client) AddParameterSubscription(topic string, parameter string) {
	if !c.ensureSubscription(topic) {
		return
	}

	c.subscriptions[topic].parameters[parameter] = true
}

func (c *client) contemplateUnsubscription(topic string) bool {
	if c.mqttClient == nil {
		return false
	}

	if len(c.subscriptions[topic].parameters) == 0 && len(c.subscriptions[topic].rules) == 0 {
		delete(c.subscriptions, topic)
		if token := c.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log.Errorln("Failed to remove subscription [%s]: %v", topic, token.Error())
			return false
		}
		log.Infof("Unsubscribed from MQTT topic [%s]", topic)
	}
	return true
}

func (c *client) RemoveParameterSubscription(topic string, parameter string) {
	_, exists := c.subscriptions[topic]
	if c.mqttClient == nil || !exists {
		return
	}

	delete(c.subscriptions[topic].parameters, parameter)

	c.contemplateUnsubscription(topic)
}

func (c *client) handleIncomingParam(param string, value string) {
	c.SetParameter(param, value)
}

func (c *client) handleIncomingRule(ruleset string, rule string, value string) {
	log.Infof("Received rule '%s/%s'", ruleset, rule)

	var r Rule
	err := json.Unmarshal([]byte(value), &r)
	if err != nil {
		log.Errorf("Unable to parse JSON string: %s", err)
		return
	}

	//TODO

}

func (c *client) Disconnect() {
	log.Infoln("Disconnecting from MQTT broker")
	c.mqttClient.Disconnect(250)
}

func (c *client) SetPrefix(prefix string) {
	c.prefix = prefix

	c.regexParam = regexp.MustCompile(fmt.Sprintf("^%sparam/([^/]+)", c.prefix))
	c.regexRule = regexp.MustCompile(fmt.Sprintf("^%srule/([^/]+)/([^/]+)", c.prefix))
}

func (c *client) GetPrefix() string {
	return c.prefix
}
