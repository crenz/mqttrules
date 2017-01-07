package agent

import (
	"testing"

	"github.com/crenz/mqttrules/test"
)

func TestAddRuleSubscription(t *testing.T) {
	testClient := New(test.NewClient(), "")

	for _, c := range []struct {
		topic  string
		result bool
	}{
		{"", false},
		{"test", true},
	} {
		testClient.AddRuleSubscription(c.topic, "ruleset", "rule")
		result := testClient.IsSubscribed(c.topic)
		if result != c.result {
			t.Errorf("IsSubscribed(%q, ...) == %v, want %v", c.topic, result, c.result)
		}
	}
}

func TestRemoveRuleSubscription(t *testing.T) {
	mqttClient := test.NewClient()
	testClient := New(mqttClient, "")

	topic := "topic"
	ruleset := "ruleset"
	rule := "rule"

	testClient.AddRuleSubscription(topic, ruleset, rule)
	result := testClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Failed to add rule subscription")
	}
	result = mqttClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Failed to add rule subscription in broker")
	}
	testClient.RemoveRuleSubscription(topic, ruleset, rule)
	result = testClient.IsSubscribed(topic)
	if result {
		t.Errorf("Failed to remove rule subscription")
	}
	result = mqttClient.IsSubscribed(topic)
	if result {
		t.Errorf("Failed to remove rule subscription in broker")
	}
	testClient.AddRuleSubscription(topic, ruleset, rule)
	testClient.AddRuleSubscription(topic, ruleset, "another_rule")
	testClient.RemoveRuleSubscription(topic, ruleset, rule)
	result = testClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Rule subscription was removed even though another rule still needs it")
	}
}
