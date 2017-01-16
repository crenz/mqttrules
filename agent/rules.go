package agent

import (
	"encoding/json"

	"fmt"
	"regexp"

	"github.com/Knetic/govaluate"
	log "github.com/Sirupsen/logrus"
	"github.com/oliveagle/jsonpath"
	"github.com/robfig/cron"
)

type Action struct {
	Topic   string
	Payload string
	QoS     byte
	Retain  bool
}

type Rule struct {
	Trigger             string
	Schedule            string
	Condition           string
	Actions             []Action
	conditionExpression *govaluate.EvaluableExpression
	cron                *cron.Cron
}

func (a *agent) GetRule(ruleset string, rule string) *Rule {
	a.rulesMutex.Lock()
	r, exists := a.rules[rulesKey{ruleset, rule}]
	a.rulesMutex.Unlock()

	if exists {
		return &r
	}
	return nil
}

func (a *agent) AddRuleFromString(ruleset string, rule string, value string) {
	log.Debugf("Received rule '%s/%s'", ruleset, rule)

	var r Rule
	err := json.Unmarshal([]byte(value), &r)
	if err != nil {
		log.Errorf("[Rule] Unable to parse JSON string: %v", err)
		return
	}

	a.AddRule(ruleset, rule, r)
}

func (a *agent) AddRule(ruleset string, rule string, r Rule) {
	var err error

	functions := map[string]govaluate.ExpressionFunction{
		"payload": func(args ...interface{}) (interface{}, error) { return nil, nil },
	}

	if len(r.Actions) == 0 {
		log.Errorf("Failed to add Rule that does not contain any actions")
		return
	}

	if len(r.Schedule) > 0 {
		r.cron = cron.New()
		r.cron.AddFunc(r.Schedule, func() {
			a.ExecuteRule(ruleset, rule, "")
		})
	}

	if len(r.Condition) > 0 {
		r.conditionExpression, err = govaluate.NewEvaluableExpressionWithFunctions(r.Condition, functions)
		if err != nil {
			log.Errorf("Error parsing rule condition: %v", err)
			return
		}
	}

	rk := rulesKey{ruleset, rule}

	if prevR := a.GetRule(ruleset, rule); prevR != nil {
		if len(prevR.Trigger) > 0 {
			a.RemoveRuleSubscription(r.Trigger, ruleset, rule)
		}
		if prevR.cron != nil {
			prevR.cron.Stop()
		}
	}

	a.rulesMutex.Lock()
	a.rules[rk] = r
	a.rulesMutex.Unlock()

	if len(r.Trigger) > 0 {
		a.AddRuleSubscription(r.Trigger, ruleset, rule)
	}
	if r.cron != nil {
		r.cron.Start()
	}
	log.Debugf("Added rule %s: %+v\n", rule, r)

	//TODO

}

/* public functions */

func (a *agent) AddRuleSubscription(topic string, ruleset string, rule string) {
	if !a.ensureSubscription(topic) {
		return
	}

	a.subscriptions[topic].rules[rulesKey{ruleset, rule}] = true
}

func (a *agent) RemoveRuleSubscription(topic string, ruleset string, rule string) {
	_, exists := a.subscriptions[topic]
	if a.mqttClient == nil || !exists {
		return
	}

	delete(a.subscriptions[topic].rules, rulesKey{ruleset, rule})
	a.contemplateUnsubscription(topic)
}

func (a *agent) ExecuteRule(ruleset string, rule string, triggerPayload string) {
	log.WithFields(log.Fields{
		"component": "Rules",
		"ruleset":   ruleset,
		"rule":      rule,
	}).Debug("Incoming rule execution request")

	fPayload := func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			// No JSON path given - return whole payload
			return a.parseParameterValue(triggerPayload), nil
		}
		var jsonData interface{}
		err := json.Unmarshal([]byte(triggerPayload), &jsonData)
		if err != nil {
			log.Errorf("JSON parsing error in trigger payload when executing rule %s/%s: %v", ruleset, rule, err)
			return triggerPayload, err
		}
		res, err := jsonpath.JsonPathLookup(jsonData, args[0].(string))
		if err != nil {
			log.Errorf("JSON lookup error in trigger payload when executing rule %s/%s: %v", ruleset, rule, err)
			return triggerPayload, err
		}

		return res, nil
	}

	functions := map[string]govaluate.ExpressionFunction{
		"payload": fPayload,
	}

	r := a.GetRule(ruleset, rule)
	if r == nil {
		return
	}
	if len(r.Condition) > 0 {
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(r.Condition, functions)
		if err != nil {
			log.Errorln("Error parsing condition:", err)
			return
		}
		result, err := expression.Evaluate(a.parameterValues)
		if err != nil {
			log.Errorln("Error evaluating condition:", err)
			return
		}
		if result != true {
			log.Debugln("Condition evaluated to false, rule not executed")
			return
		}
	}
	for _, r := range r.Actions {
		s := a.EvalExpressionsInString(r.Payload, functions)
		a.Publish(r.Topic, r.QoS, r.Retain, s)
	}
}

func (a *agent) EvalExpressionsInString(in string, functions map[string]govaluate.ExpressionFunction) string {
	r := regexp.MustCompile("[$][{].*?[}]")
	out := r.ReplaceAllStringFunc(in, func(i string) string {
		// ReplaceAllStringFunc always receives the complete match, cannot receive
		// submatches -> therefore, we chomp first two and last character off in this
		// hackish way
		e := i[2 : len(i)-1]
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(e, functions)
		if err != nil {
			log.Errorln("Error parsing expression:", err)
			log.Errorln("String: ", in, "; Expression:", e, "; i: ", i)
			return ""
		}
		a.paramMutex.Lock()
		v := a.parameterValues
		a.paramMutex.Unlock()
		result, err := expression.Evaluate(v)
		if err != nil {
			log.Errorln("Error evaluating expression:", err)
			return ""
		}

		return fmt.Sprintf("%v", result)
	})
	return out
}
