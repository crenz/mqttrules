package mqttrules

import (
	"encoding/json"
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/oliveagle/jsonpath"
)

// Parameter used in MQTT rules; can be updated from incoming MQTT messages
type Parameter struct {
	Value    string
	Topic    string
	JsonPath string
}

type parameterMap map[string]*Parameter

func (c *client) SetParameter(parameter string, value string) {
	if len(parameter) == 0 {
		return
	}

	var p Parameter
	err := json.Unmarshal([]byte(value), &p)
	if err != nil {
		c.parameters[parameter] = &Parameter{value, "", ""}
		log.Infof("Setting parameter %s to non-JSON value", parameter)
		return
	}
	if c.parameters[parameter] != nil && len(c.parameters[parameter].Topic) > 0 {
		c.RemoveParameterSubscription(p.Topic, parameter)
	}
	c.parameters[parameter] = &p
	if len(c.parameters[parameter].Topic) > 0 {
		c.AddParameterSubscription(p.Topic, parameter)
	}
	log.Infof("Setting parameter %s to JSON value %s", parameter, fmt.Sprintf("%+v\n", p))

}

func (c *client) TriggerParameterUpdate(parameter string, value string) {
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
		c.parameters[parameter].Value = fmt.Sprintf("%v", res)
		log.Infof("Updated parameter %s to value %s", parameter, c.parameters[parameter].Value)
	}

}

func (c *client) GetParameterValue(parameter string) string {
	if c.parameters[parameter] == nil {
		return ""
	}

	return c.parameters[parameter].Value
}

func (c *client) ReplaceParamsInString(in string) string {
	r := regexp.MustCompile("[$].*[$]")
	out := r.ReplaceAllStringFunc(in, func(i string) string {
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first and last character off in this
		// hackish way
		if len(i) == 2 {
			// '$$' -> '$'
			return "$"
		}

		return c.GetParameterValue(i[1 : len(i)-1])
	})
	return out
}
