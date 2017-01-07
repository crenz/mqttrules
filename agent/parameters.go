package agent

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/oliveagle/jsonpath"
)

// Parameter used in MQTT rules; can be updated from incoming MQTT messages
type Parameter struct {
	Value    interface{}
	Topic    string
	JsonPath string
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
	p, exists := c.parameters[parameter]
	if !exists {
		return
	}

	if len(p.JsonPath) == 0 {
		// directly set value
		c.parameters[parameter].Value = value
		log.Infof("Updated parameter %s to non-JSON value %s", parameter, fmt.Sprintf("%+v\n", p))
	} else {
		var jsonData interface{}
		err := json.Unmarshal([]byte(value), &jsonData)
		if err != nil {
			log.Errorf("JSON parsing error when updating parameter %s: %v", parameter, err)
			return
		}
		res, err := jsonpath.JsonPathLookup(jsonData, p.JsonPath)
		if err != nil {
			log.Errorf("JSON error when updating parameter %s: %v", parameter, err)
			return
		}
		c.SetParameterValue(parameter, res)
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

func (c *agent) EvalExpressionsInString(in string, functions map[string]govaluate.ExpressionFunction) string {
	r := regexp.MustCompile("[$][{].*[}]")
	out := r.ReplaceAllStringFunc(in, func(i string) string {
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first two and last character off in this
		// hackish way
		e := i[2 : len(i)-1]
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(e, functions)
		if err != nil {
			log.Errorln("Error parsing expression:", err)
			return ""
		}
		result, err := expression.Evaluate(c.parameterValues)
		if err != nil {
			log.Errorln("Error evaluating expression:", err)
			return ""
		}

		return fmt.Sprintf("%v", result)
	})
	return out
}

func (c *agent) ReplaceParamsInString(in string) string {
	r := regexp.MustCompile("[$].*[$]")
	out := r.ReplaceAllStringFunc(in, func(i string) string {
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first and last character off in this
		// hackish way
		if len(i) == 2 {
			// '$$' -> '$'
			return "$"
		}

		return c.GetParameterValue(i[1 : len(i)-1]).(string)
	})
	return out
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
