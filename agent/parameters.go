package agent

import (
	"encoding/json"

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
	a.paramMutex.Lock()
	a.parameters[name] = &p
	a.paramMutex.Unlock()
	a.SetParameterValue(name, p.Value)
	if len(a.parameters[name].Topic) > 0 {
		a.AddParameterSubscription(p.Topic, name)
	}
	log.Debugf("Setting parameter %s to JSON value %+v\n", name, p)
}

func (a *agent) SetParameterValue(parameter string, value interface{}) {
	a.paramMutex.Lock()
	a.parameterValues[parameter] = value
	a.paramMutex.Unlock()
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

	a.paramMutex.Lock()
	p, exists := a.parameters[parameter]
	a.paramMutex.Unlock()
	if !exists {
		return
	}

	if len(p.Expression) == 0 {
		// directly set value
		a.SetParameterValue(parameter, value)
		log.WithFields(log.Fields{
			"component": "Parameters",
			"parameter": parameter,
			"value":     value,
			"mode":      "fullPayload",
		}).Debug("Parameter value updated")
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
		log.WithFields(log.Fields{
			"component": "Parameters",
			"parameter": parameter,
			"mode":      "JSON",
			"value":     result,
		}).Debug("Parameter value updated")
	}

}

func (a *agent) GetParameterValue(parameter string) interface{} {
	a.paramMutex.Lock()
	v, exists := a.parameterValues[parameter]
	a.paramMutex.Unlock()
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
