package mqttrules

import (
	"regexp"
)

type action struct {
	ID string
	Value string
	Retain bool
	QoS int
}

type rule struct {
	Conditions[] string
	Actions[] action
}

var parameters = map[string]string{}
var rules = map[string]rule{}

/* public functions */

func SetParameter(parameter string, value string) {
	parameters[parameter] = value
}

func GetParameter(parameter string) string {
	return parameters[parameter]
}

func ReplaceParamsInString(in string) string {
	r := regexp.MustCompile("[$].*[$]")
	out := r.ReplaceAllStringFunc(in, func(i string)string{
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first and last character off in this
		// hackish way
		if (len(i) == 2) {
			// '$$' -> '$'
			return "$"
		} else {
			return GetParameter(i[1:len(i)-1])
		}
	})
	return out
}

func SetRule(id string, conditions[] string, actions[] action) {
	rules[id] = rule{conditions, actions}
}

func GetRule(id string) rule {
	return rules[id]
}

func ExecuteRule(id string) action {
	return action{}
}

