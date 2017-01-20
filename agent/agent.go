package agent

import (
	"fmt"
	"regexp"

	"strings"

	"encoding/json"

	"sync"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
)

// MqttClient provides the interface to a MQTT Client implementation. This interface is provided to enable
// tests with a mock client
type MqttClient interface {
	IsConnected() bool
	Connect() bool
	Disconnect()
	SetSubscriptionCallback(callback func(string, string))
	Publish(topic string, qos byte, retained bool, payload interface{}) bool
	Subscribe(topic string, qos byte) bool
	Unsubscribe(topics ...string) bool
}

type MessageHandler func(topic string, payload string)

// Agent implements the main rules agent
type Agent interface {
	Connect() bool
	Subscribe() bool
	Listen()
	Disconnect()

	HandleMessage(topic string, payload []byte)

	SetParameterFromString(name string, value string)
	SetParameter(name string, param Parameter)
	GetParameterValue(parameter string) interface{}
	TriggerParameterUpdate(parameter string, value string)
	EvalExpressionsInString(in string, functions map[string]govaluate.ExpressionFunction) string

	AddParameterSubscription(topic string, parameter string)
	RemoveParameterSubscription(topic string, parameter string)

	AddRuleFromString(ruleset string, rule string, value string)
	AddRule(ruleset string, rule string, r Rule)
	GetRule(ruleset string, rule string) *Rule
	AddRuleSubscription(topic string, ruleset string, rule string)
	RemoveRuleSubscription(topic string, ruleset string, rule string)

	ExecuteRule(ruleset string, rule string, triggerPayload string)
	Publish(topic string, qos byte, retained bool, payload string)
	IsSubscribed(topic string) bool
	InjectConfigFile(c ConfigFile)
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
	regexSys       *regexp.Regexp
	messagehandler MessageHandler

	rulesMutex sync.Mutex
	paramMutex sync.Mutex
}

func (a *agent) initialize() {
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
	mqttClient.SetSubscriptionCallback(a.messagehandler)
	a.mqttClient = mqttClient
	a.setPrefix(prefix)

	return a
}

func (a *agent) Connect() bool {
	return a.mqttClient.Connect()
}

func (a *agent) Subscribe() bool {
	return a.mqttClient.Subscribe(fmt.Sprintf("%sparam/+", a.prefix), byte(1)) &&
		a.mqttClient.Subscribe(fmt.Sprintf("%srule/+/+", a.prefix), byte(1)) &&
		a.mqttClient.Subscribe(fmt.Sprintf("%s$MQTTRULES", a.prefix), byte(1))
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
		a.SetParameterFromString(res[1], string(payload))
	}
	if res := a.regexRule.FindStringSubmatch(topic); res != nil {
		a.AddRuleFromString(res[1], res[2], string(payload))
	}
	if a.regexSys.MatchString(topic) {
		switch {
		case strings.Compare("parameters", string(payload)) == 0:
			for key := range a.parameters {
				s, _ := json.Marshal(a.parameters[key])
				a.Publish(fmt.Sprintf("%s$MQTTRULES/parameters/%s", a.prefix, key), 2, false, string(s))
			}
		case strings.Compare("rules", string(payload)) == 0:
			for rk := range a.rules {
				s, _ := json.Marshal(a.rules[rk])
				a.Publish(fmt.Sprintf("%s$MQTTRULES/rules/%s/%s", a.prefix, rk.ruleset, rk.rule), 2, false, string(s))
			}
			a.Publish(fmt.Sprintf("%s$MQTTRULES/rules", a.prefix), 2, false, fmt.Sprintf("%+v", a.rules))
		default:
			a.Publish(fmt.Sprintf("%s$MQTTRULES", a.prefix), 2, false, fmt.Sprintf("Unknown command '%s'", string(payload)))
		}
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
		if success := a.mqttClient.Subscribe(topic, byte(1)); !success {
			log.Errorln("Failed to add subscription [%s]")
			return false
		}
		log.Debugf("Subscribed to MQTT topic [%s]", topic)
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
	a.regexSys = regexp.MustCompile(fmt.Sprintf("^%s[$]MQTTRULES", a.prefix))

}

func (a *agent) IsSubscribed(topic string) bool {
	_, exists := a.subscriptions[topic]
	return exists
}

func (a *agent) InjectConfigFile(c ConfigFile) {
	for n, p := range c.Parameters {
		a.SetParameter(n, p)
	}

	for ruleset := range c.Rules {
		for rule := range c.Rules[ruleset] {
			a.AddRule(ruleset, rule, c.Rules[ruleset][rule])
		}
	}
}
