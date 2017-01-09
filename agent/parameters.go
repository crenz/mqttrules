package agent

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/oliveagle/jsonpath"
)

// Parameter used in MQTT rules; can be updated from incoming MQTT messages
type Parameter struct {
	Value      interface{}
	Topic      string
	Expression string
}

type parameterMap map[string]*Parameter

func (c *agent) SetParameter(parameter string, value string) {
	if len(parameter) == 0 {
		return
	}

	var p Parameter
	err := json.Unmarshal([]byte(value), &p)
	spew.Dump(p)
	if err != nil {
		c.parameters[parameter] = &Parameter{value, "", ""}
		c.SetParameterValue(parameter, value)
		log.Infof("Setting parameter %s to non-JSON value", parameter)
		return
	}
	if c.parameters[parameter] != nil && len(c.parameters[parameter].Topic) > 0 {
		c.RemoveParameterSubscription(p.Topic, parameter)
	}
	c.parameters[parameter] = &p
	c.SetParameterValue(parameter, p.Value)
	if len(c.parameters[parameter].Topic) > 0 {
		c.AddParameterSubscription(p.Topic, parameter)
	}
	log.Infof("Setting parameter %s to JSON value %+v\n", parameter, p)

}

func (c *agent) SetParameterValue(parameter string, value interface{}) {
	c.parameterValues[parameter] = value
}

func (c *agent) TriggerParameterUpdate(parameter string, value string) {
	fPayload := func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			// No JSON path given - return whole payload
			return value, nil
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

	p, exists := c.parameters[parameter]
	if !exists {
		return
	}

	if len(p.Expression) == 0 {
		// directly set value
		c.parameters[parameter].Value = value
		log.Infof("Updated parameter %s to non-JSON value %s", parameter, fmt.Sprintf("%+v\n", p))
	} else {
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(p.Expression, functions)
		if err != nil {
			log.Errorln("Error parsing condition:", err)
			return
		}
		result, err := expression.Evaluate(c.parameterValues)
		if err != nil {
			log.Errorln("Error evaluating condition:", err)
			return
		}

		c.SetParameterValue(parameter, result)
		log.Infof("Updated parameter %s to value %s", parameter, c.parameters[parameter].Value)
	}

}

func (c *agent) GetParameterValue(parameter string) interface{} {
	v, exists := c.parameterValues[parameter]
	if !exists {
		v = ""
	}

	return v
}

func (c *agent) AddParameterSubscription(topic string, parameter string) {
	if !c.ensureSubscription(topic) {
		return
	}

	c.subscriptions[topic].parameters[parameter] = true
}

func (c *agent) RemoveParameterSubscription(topic string, parameter string) {
	_, exists := c.subscriptions[topic]
	if c.mqttClient == nil || !exists {
		return
	}

	delete(c.subscriptions[topic].parameters, parameter)

	c.contemplateUnsubscription(topic)
}
