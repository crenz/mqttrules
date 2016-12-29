package mqttrules

import (
	"regexp"
)

type Action struct {
	Topic string
	Payload string
	QoS int
	Retain bool
}

type Rule struct {
	Trigger string
	Schedule string
	Conditions[] string
	Actions[] Action
}

var parameters = map[string]string{}
var rules = map[string]Rule{}

/* public functions */

func (c *client) SetParameter(parameter string, value string) {
	c.parameters[parameter] = value
}

func (c *client) GetParameter(parameter string) string {
	return c.parameters[parameter]
}

func (c *client) ReplaceParamsInString(in string) string {
	r := regexp.MustCompile("[$].*[$]")
	out := r.ReplaceAllStringFunc(in, func(i string)string{
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first and last character off in this
		// hackish way
		if (len(i) == 2) {
			// '$$' -> '$'
			return "$"
		} else {
			return c.GetParameter(i[1:len(i)-1])
		}
	})
	return out
}

func (c *client) SetRule(id string, conditions[] string, actions[] Action) {
//	rules[id] = Rule{conditions, actions}
}

func (c *client) GetRule(id string) Rule {
	return rules[id]
}

func (c *client) ExecuteRule(id string) Action {
	return Action{}
}

