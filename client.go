package mqttrules

import (
	log "github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"regexp"
	"fmt"
//	"encoding/json"
	"encoding/json"
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
	GetParameter(parameter string) string
	ReplaceParamsInString(in string) string
	SetRule(id string, conditions[] string, actions[] Action)
	GetRule(id string) Rule
	ExecuteRule(id string) Action
}

type paramMap map[string]string
type rulesMap map[string]map[string]Rule

type client struct {
	mqttClient mqtt.Client
	messages chan[2]string
	prefix string

	parameters paramMap
	rules rulesMap

	regexParam * regexp.Regexp
	regexRule * regexp.Regexp
}

func (c *client) initialize() {
	c.SetPrefix("")
	c.parameters = make(paramMap)
	c. rules = make(rulesMap)
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
	if (token.Error() != nil) {
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
	incoming := <- c.messages
	//log.Infof("Received [%s] %s\n", incoming[0], incoming[1])
	if res := c.regexParam.FindStringSubmatch(incoming[0]); res != nil {
		c.handleIncomingParam(res[1], incoming[1])
	}
	if res := c.regexRule.FindStringSubmatch(incoming[0]); res != nil {
		c.handleIncomingRule(res[1], res[2], incoming[1])
	}
}

func (c *client) handleIncomingParam(param string, value string) {
	log.Infof("Setting parameter '%s'", param)
	c.SetParameter(param, value)
}

func (c *client) handleIncomingRule(ruleset string, rule string, value string) {
	log.Infof("Received rule '%s/%s'", ruleset, rule)

	var r Rule
	err := json.Unmarshal([]byte(value), &r)
	if (err != nil) {
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
