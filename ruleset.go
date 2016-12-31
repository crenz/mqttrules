package mqttrules

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

var rules = map[string]Rule{}

/* public functions */

func (c *client) SetRule(id string, conditions[] string, actions[] Action) {
//	rules[id] = Rule{conditions, actions}
}

func (c *client) GetRule(id string) Rule {
	return rules[id]
}

func (c *client) ExecuteRule(id string) Action {
	return Action{}
}

