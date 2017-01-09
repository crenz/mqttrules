package agent

import (
	"fmt"
	"regexp"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
)

// Interface for MQTT client used to interface with broker
type MqttClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte, callback func(string, string)) bool
	Unsubscribe(topics ...string) bool
}

type MessageHandler func(topic string, payload string)

// MQTT Rules agent
type Agent interface {
	Connect() bool
	Subscribe() bool
	Listen()
	Disconnect()

	HandleMessage(topic string, payload []byte)

	SetParameter(parameter string, value string)
	GetParameterValue(parameter string) interface{}
	TriggerParameterUpdate(parameter string, value string)
	EvalExpressionsInString(in string, functions map[string]govaluate.ExpressionFunction) string

	AddParameterSubscription(topic string, parameter string)
	RemoveParameterSubscription(topic string, parameter string)

	AddRule(ruleset string, rule string, value string)
	GetRule(ruleset string, rule string) *Rule
	AddRuleSubscription(topic string, ruleset string, rule string)
	RemoveRuleSubscription(topic string, ruleset string, rule string)

	ExecuteRule(ruleset string, rule string, triggerPayload string)
	Publish(topic string, qos byte, retained bool, payload string)
	IsSubscribed(topic string) bool
}

type rulesKey struct {
	ruleset, rule string
}

type rulesMap map[rulesKey]Rule

type subscriptions struct {
	//TODO Use slices instead of maps
	parameters map[string]bool
	rules      map[rulesKey]bool
}

type subscriptionsMap map[string]subscriptions

type agent struct {
	mqttClient MqttClient
	messages   chan [2]string
	prefix     string

	parameters      parameterMap
	parameterValues map[string]interface{}
	rules           rulesMap
	subscriptions   subscriptionsMap

	regexParam     *regexp.Regexp
	regexRule      *regexp.Regexp
	messagehandler MessageHandler
}

func (a *agent) initialize() {
	a.setPrefix("")
	a.parameters = make(parameterMap)
	a.parameterValues = make(map[string]interface{})
	a.rules = make(rulesMap)
	a.subscriptions = make(subscriptionsMap)
	a.messagehandler = func(topic string, payload string) {
		a.messages <- [2]string{topic, payload}
	}
	a.messages = make(chan [2]string)
}

// Creates and initializes a new MQTT rules client
func New(mqttClient MqttClient, prefix string) Agent {
	a := &agent{}
	a.initialize()
	a.mqttClient = mqttClient
	a.prefix = prefix

	return a
}

func (a *agent) Connect() bool {
	log.Infoln("Connecting to MQTT broker")

	return a.mqttClient.Connect()
}

func (a *agent) Subscribe() bool {
	return a.mqttClient.Subscribe("#", byte(1), a.messagehandler)
}

func (a *agent) Publish(topic string, qos byte, retained bool, payload string) {

	if success := a.mqttClient.Publish(topic, qos, retained, payload); !success {
		log.Errorf("Error publishing MQTT topic [%s]", topic)
	}
}

func (a *agent) HandleMessage(topic string, payload []byte) {
	_, exists := a.subscriptions[topic]
	if exists {
		a.handleIncomingTrigger(topic, string(payload))
	}

	if res := a.regexParam.FindStringSubmatch(topic); res != nil {
		a.SetParameter(res[1], string(payload))
	}
	if res := a.regexRule.FindStringSubmatch(topic); res != nil {
		a.AddRule(res[1], res[2], string(payload))
	}

}

func (a *agent) Listen() {
	incoming := <-a.messages
	//log.Infof("Received [%s] %s\n", incoming[0], incoming[1])
	a.HandleMessage(incoming[0], []byte(incoming[1]))

}

func (a *agent) handleIncomingTrigger(topic string, payload string) {
	for key := range a.subscriptions[topic].parameters {
		a.TriggerParameterUpdate(key, payload)
	}
	for key := range a.subscriptions[topic].rules {
		a.ExecuteRule(key.ruleset, key.rule, payload)
	}
}

func (a *agent) ensureSubscription(topic string) bool {
	if a.mqttClient == nil || len(topic) == 0 {
		return false
	}

	if _, exists := a.subscriptions[topic]; !exists {
		a.subscriptions[topic] = subscriptions{parameters: make(map[string]bool), rules: make(map[rulesKey]bool)}
	}
	if len(a.subscriptions[topic].parameters) == 0 && len(a.subscriptions[topic].rules) == 0 {
		if success := a.mqttClient.Subscribe(topic, byte(1), a.messagehandler); !success {
			log.Errorln("Failed to add subscription [%s]")
			return false
		}
		log.Infof("Subscribed to MQTT topic [%s]", topic)
	}
	return true
}

func (a *agent) contemplateUnsubscription(topic string) bool {
	if a.mqttClient == nil {
		return false
	}

	if len(a.subscriptions[topic].parameters) == 0 && len(a.subscriptions[topic].rules) == 0 {
		delete(a.subscriptions, topic)
		if success := a.mqttClient.Unsubscribe(topic); !success {
			log.Errorln("Failed to remove subscription [%s]")
			return false
		}
		log.Infof("Unsubscribed from MQTT topic [%s]", topic)
	}
	return true
}

func (a *agent) Disconnect() {
	log.Infoln("Disconnecting from MQTT broker")
	a.mqttClient.Disconnect()
}

func (a *agent) setPrefix(prefix string) {
	a.prefix = prefix

	a.regexParam = regexp.MustCompile(fmt.Sprintf("^%sparam/([^/]+)", a.prefix))
	a.regexRule = regexp.MustCompile(fmt.Sprintf("^%srule/([^/]+)/([^/]+)", a.prefix))
}

func (a *agent) GetPrefix() string {
	return a.prefix
}

func (a *agent) IsSubscribed(topic string) bool {
	_, exists := a.subscriptions[topic]
	return exists
}
