package agent

import (
	"encoding/json"
	"fmt"

	"strconv"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
	"github.com/oliveagle/jsonpath"
)

// Parameter used in MQTT rules; can be updated from incoming MQTT messages
type Parameter struct {
	Value      interface{}
	Topic      string
	Expression string
}

type parameterMap map[string]*Parameter

func (a *agent) parseParameterValue(value string) interface{} {
	f, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return f
	}

	return value
}

func (a *agent) SetParameterFromString(name string, value string) {
	if len(name) == 0 {
		return
	}

	var p Parameter
	err := json.Unmarshal([]byte(value), &p)
	if err != nil {
		v := a.parseParameterValue(value)
		a.parameters[name] = &Parameter{v, "", ""}
		a.SetParameterValue(name, v)
		log.Debugf("Setting parameter %s to non-JSON value", name)
		return
	}
	a.SetParameter(name, p)
}

func (a *agent) SetParameter(name string, p Parameter) {
	if a.parameters[name] != nil && len(a.parameters[name].Topic) > 0 {
		a.RemoveParameterSubscription(p.Topic, name)
	}
	a.parameters[name] = &p
	a.SetParameterValue(name, p.Value)
	if len(a.parameters[name].Topic) > 0 {
		a.AddParameterSubscription(p.Topic, name)
	}
	log.Debugf("Setting parameter %s to JSON value %+v\n", name, p)
}

func (a *agent) SetParameterValue(parameter string, value interface{}) {
	a.parameterValues[parameter] = value
}

func (a *agent) TriggerParameterUpdate(parameter string, value string) {
	fPayload := func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			// No JSON path given - return whole payload.
			return a.parseParameterValue(value), nil
		}
		var jsonData interface{}
		err := json.Unmarshal([]byte(value), &jsonData)
		if err != nil {
			log.Errorf("JSON parsing error in trigger payload when updating parameter %s: %v", parameter, err)
			return value, err
		}
		res, err := jsonpath.JsonPathLookup(jsonData, args[0].(string))
		if err != nil {
			log.Errorf("JSON lookup error in trigger payload when updating parameter %s: %v", parameter, err)
			return value, err
		}

		return res, nil
	}

	functions := map[string]govaluate.ExpressionFunction{
		"payload": fPayload,
	}

	p, exists := a.parameters[parameter]
	if !exists {
		return
	}

	if len(p.Expression) == 0 {
		// directly set value
		a.parameters[parameter].Value = value
		log.Debugf("Updated parameter %s to non-JSON value %s", parameter, fmt.Sprintf("%+v\n", p))
	} else {
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(p.Expression, functions)
		if err != nil {
			log.Errorln("Error parsing parameter expression :", err)
			return
		}
		result, err := expression.Evaluate(a.parameterValues)
		if err != nil {
			log.Errorln("Error evaluating parameter expression:", err)
			return
		}

		a.SetParameterValue(parameter, result)
		log.Debugf("Updated parameter %s to value %s", parameter, a.parameters[parameter].Value)
	}

}

func (c *agent) GetParameterValue(parameter string) interface{} {
	v, exists := c.parameterValues[parameter]
	if !exists {
		v = ""
	}

	return v
}

func (a *agent) AddParameterSubscription(topic string, parameter string) {
	if !a.ensureSubscription(topic) {
		return
	}

	a.subscriptions[topic].parameters[parameter] = true
}

func (a *agent) RemoveParameterSubscription(topic string, parameter string) {
	_, exists := a.subscriptions[topic]
	if a.mqttClient == nil || !exists {
		return
	}

	delete(a.subscriptions[topic].parameters, parameter)

	a.contemplateUnsubscription(topic)
}
