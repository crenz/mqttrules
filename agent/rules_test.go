package agent

import (
	"testing"

	"strings"

	"github.com/crenz/mqttrules/test"
	"github.com/davecgh/go-spew/spew"
)

func TestAgent_AddRuleSubscription(t *testing.T) {
	a := New(test.NewClient(), "")

	for _, c := range []struct {
		topic  string
		result bool
	}{
		{"", false},
		{"test", true},
	} {
		a.AddRuleSubscription(c.topic, "ruleset", "rule")
		result := a.IsSubscribed(c.topic)
		if result != c.result {
			t.Errorf("IsSubscribed(%q, ...) == %v, want %v", c.topic, result, c.result)
		}
	}
}

func TestAgent_RemoveRuleSubscription(t *testing.T) {
	mqttClient := test.NewClient()
	a := New(mqttClient, "")

	topic := "topic"
	ruleset := "ruleset"
	rule := "rule"

	a.AddRuleSubscription(topic, ruleset, rule)
	result := a.IsSubscribed(topic)
	if !result {
		t.Errorf("Failed to add rule subscription")
	}
	result = mqttClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Failed to add rule subscription in broker")
	}
	a.RemoveRuleSubscription(topic, ruleset, rule)
	result = a.IsSubscribed(topic)
	if result {
		t.Errorf("Failed to remove rule subscription")
	}
	result = mqttClient.IsSubscribed(topic)
	if result {
		t.Errorf("Failed to remove rule subscription in broker")
	}
	a.AddRuleSubscription(topic, ruleset, rule)
	a.AddRuleSubscription(topic, ruleset, "another_rule")
	a.RemoveRuleSubscription(topic, ruleset, rule)
	result = a.IsSubscribed(topic)
	if !result {
		t.Errorf("Rule subscription was removed even though another rule still needs it")
	}
}

func TestAgent_AddRule(t *testing.T) {
	mqttClient := test.NewClient()
	a := New(mqttClient, "")

	ruleset := "ruleset"
	rule := "rule"

	/* Invalid rules */
	a.AddRule(ruleset, rule, "")
	r := a.GetRule(ruleset, rule)
	if r != nil {
		t.Errorf("Rule should not have been added: Empty specification")
	}
	a.AddRule(ruleset, rule, `invalid-json`)
	r = a.GetRule(ruleset, rule)
	if r != nil {
		t.Errorf("Rule should not have been added: Invalid JSON")
	}
	a.AddRule(ruleset, rule, `{"trigger": "test"}`)
	r = a.GetRule(ruleset, rule)
	if r != nil {
		t.Errorf("Rule should not have been added: No actions")
	}
	a.AddRule(ruleset, rule, `{"trigger": "test", "actions": []}`)
	r = a.GetRule(ruleset, rule)
	if r != nil {
		t.Errorf("Rule should not have been added: No actions")
	}

	/* Simple rules */
	a.AddRule(ruleset, rule, `{"trigger": "test", "actions": [
          {
            "topic": "send_topic",
            "payload": "send_payload",
            "qos": 2,
            "retain": false
          }
	]}`)
	r = a.GetRule(ruleset, rule)
	if r == nil || strings.Compare(r.Actions[0].Topic, "send_topic") != 0 ||
		strings.Compare(r.Actions[0].Payload, "send_payload") != 0 ||
		r.Actions[0].QoS != 2 || r.Actions[0].Retain != false ||
		len(r.Condition) != 0 || r.cron != nil {
		t.Errorf("Failed to add rule with correct data")
	}
	a.AddRule(ruleset, rule, `{"trigger": "test", "actions": [
          {
            "topic": "send/complex/topic",
            "qos": 1,
            "retain": true
          }
	]}`)
	r = a.GetRule(ruleset, rule)
	if r == nil || strings.Compare(r.Actions[0].Topic, "send/complex/topic") != 0 ||
		len(r.Actions[0].Payload) != 0 ||
		r.Actions[0].QoS != 1 || r.Actions[0].Retain != true ||
		len(r.Condition) != 0 || r.cron != nil {
		t.Errorf("Failed to add rule with correct data")
	}

	/* Rules with conditions */
	a.AddRule(ruleset, rule, `{
		"trigger": "test",
	        "condition": "1 > 0",
		"actions": [{
			"topic": "send/conditional",
	        	"qos": 1,
        	 	"retain": true
		}]}`)
	r = a.GetRule(ruleset, rule)
	if r == nil || strings.Compare(r.Actions[0].Topic, "send/conditional") != 0 ||
		len(r.Actions[0].Payload) != 0 ||
		r.Actions[0].QoS != 1 || r.Actions[0].Retain != true ||
		strings.Compare(r.Condition, "1 > 0") != 0 || r.conditionExpression == nil || r.cron != nil {
		t.Errorf("Failed to add rule with correct data")
		spew.Dump(r)
	}

	/* Rules with schedule */
	a.AddRule(ruleset, rule, `{
		"trigger": "test",
	        "condition": "1 > 0",
	        "schedule": "@every 10s",
		"actions": [{
			"topic": "send/conditional",
	        	"qos": 1,
        	 	"retain": true
		}]}`)
	r = a.GetRule(ruleset, rule)
	if r == nil || strings.Compare(r.Actions[0].Topic, "send/conditional") != 0 ||
		len(r.Actions[0].Payload) != 0 ||
		r.Actions[0].QoS != 1 || r.Actions[0].Retain != true ||
		strings.Compare(r.Condition, "1 > 0") != 0 || r.conditionExpression == nil ||
		strings.Compare(r.Schedule, "@every 10s") != 0 || r.cron == nil {
		t.Errorf("Failed to add rule with correct data")
		spew.Dump(r)
	}
}

func TestAgent_ExecuteRule(t *testing.T) {
	mqttClient := test.NewClient()
	a := New(mqttClient, "")

	ruleset := "ruleset"
	rule := "rule"

	a.AddRule(ruleset, rule, `{"trigger": "test", "actions": [
          {
            "topic": "send_topic",
            "payload": "send_payload",
            "qos": 2,
            "retain": false
          }
	]}`)
	a.ExecuteRule(ruleset, rule, "")
	m := mqttClient.LastMessage()
	if strings.Compare(m.Topic, "send_topic") != 0 ||
		strings.Compare(m.Payload.(string), "send_payload") != 0 ||
		m.QoS != 2 || m.Retained != false {
		t.Errorf("Failed to execute rule correctly")
		spew.Dump(m)
	}

	a.AddRule(ruleset, rule, `{"trigger": "test", "condition": "param > 41", "actions": [
          {
            "topic": "condition_test",
            "payload": "send_payload",
            "qos": 2,
            "retain": false
          }
	]}`)
	a.ExecuteRule(ruleset, rule, "")
	m = mqttClient.LastMessage()
	if strings.Compare(m.Topic, "condition_test") == 0 {
		t.Errorf("Rule should not have been executed: Parameter in condition not defined")
	}
	a.SetParameter("param", `{"value": 42}`)
	a.ExecuteRule(ruleset, rule, "")
	m = mqttClient.LastMessage()
	if strings.Compare(m.Topic, "condition_test") != 0 ||
		strings.Compare(m.Payload.(string), "send_payload") != 0 ||
		m.QoS != 2 || m.Retained != false {
		t.Errorf("Failed to execute rule correctly")
		spew.Dump(m)
	}

	a.AddRule(ruleset, rule, `{"trigger": "test", "condition": "param > 41", "actions": [
          {
            "topic": "condition_test",
            "payload": "${param}",
            "qos": 1,
            "retain": true
          }
	]}`)
	a.ExecuteRule(ruleset, rule, "")
	m = mqttClient.LastMessage()
	if strings.Compare(m.Topic, "condition_test") != 0 ||
		strings.Compare(m.Payload.(string), "42") != 0 ||
		m.QoS != 1 || m.Retained != true {
		t.Errorf("Failed to execute rule correctly")
		spew.Dump(m)
	}

}
