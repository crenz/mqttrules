package agent

import (
	"testing"

	"fmt"

	"github.com/crenz/mqttrules/test"
	"github.com/davecgh/go-spew/spew"
)

func TestSetParameter(t *testing.T) {
	testClient := New(test.NewClient(), "")

	for _, c := range []struct {
		key, value string
	}{
		{"/rules/TTL", "Testing"},
		{"/rules/TTL", ""},
		{"", ""},
	} {
		testClient.SetParameterFromString(c.key, c.value)
		result := testClient.GetParameterValue(c.key)
		if result != c.value {
			t.Errorf("GetParameter(%q) == %q, want %q", c.key, result, c.value)
		}
	}
}

func TestSetParameter_JSON(t *testing.T) {
	testClient := New(test.NewClient(), "")

	key := "JSONTest"

	result := testClient.GetParameterValue(key).(string)
	if len(result) != 0 {
		t.Errorf("GetParameter(%q) == %q, want empty string", key, result)
	}

	testClient.SetParameterFromString(key, `{
    "value": 42.2,
    "topic": "lighting/livingroom/status",
    "expression": "payload(\"$.value\")"
}`)
	resultFloat := testClient.GetParameterValue(key).(float64)
	if resultFloat != 42.2 {
		t.Errorf("GetParameterValue(%q) == %v, want %v", key, result, 42.2)
	}

	testClient.SetParameterFromString(key, `{
    "value": 42,
    "topic": "lighting/livingroom/status",
    "expression": "payload(\"$.value\")"
}`)
	resultFloat = testClient.GetParameterValue(key).(float64)
	if resultFloat != 42 {
		t.Errorf("GetParameterValue(%q) == %v, want %v", key, result, 42)
	}
}

func TestSetParameter_Expression(t *testing.T) {
	mqttClient := test.NewClient()
	a := New(mqttClient, "")

	key := "ExpressionTest"

	result := a.GetParameterValue(key).(string)
	if len(result) != 0 {
		t.Errorf("GetParameter(%q) == %q, want empty string", key, result)
	}

	a.SetParameterFromString(key, `{
    "topic": "lighting/livingroom/status",
    "expression": "payload(\"$.value\") + 2"
}`)
	mqttClient.Publish("lighting/livingroom/status", 1, false, `{"value": 42}`)
	a.Listen()
	result2 := a.GetParameterValue(key)
	spew.Dump(result)
	if result2.(float64) != 44 {
		t.Errorf("GetParameterValue(%q) == %v, want %v", key, result, 44)
	}
}
func TestAddParameterSubscription(t *testing.T) {
	testClient := New(test.NewClient(), "")

	for _, c := range []struct {
		topic  string
		result bool
	}{
		{"", false},
		{"test", true},
	} {
		testClient.AddParameterSubscription(c.topic, "param")
		result := testClient.IsSubscribed(c.topic)
		if result != c.result {
			t.Errorf("IsSubscribed(%q, ...) == %v, want %v", c.topic, result, c.result)
		}
	}
}

func TestRemoveParameterSubscription(t *testing.T) {
	testClient := New(test.NewClient(), "")

	topic := "topic"
	param := "param"

	testClient.AddParameterSubscription(topic, param)
	result := testClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Failed to add param subscription")
	}
	testClient.RemoveParameterSubscription(topic, param)
	result = testClient.IsSubscribed(topic)
	if result {
		t.Errorf("Failed to remove param subscription")
	}
	testClient.AddParameterSubscription(topic, param)
	testClient.AddParameterSubscription(topic, "another_param")
	testClient.RemoveParameterSubscription(topic, param)
	result = testClient.IsSubscribed(topic)
	if !result {
		t.Errorf("Param subscription was removed even though another parameter still needs it")
	}
}

func TestAgent_TriggerParameterUpdate(t *testing.T) {
	tc := New(test.NewClient(), "")

	for _, v := range []interface{}{
		float64(42),
		42.1234,
	} {
		tc.SetParameterFromString("test", fmt.Sprintf("%v", v))
		if r := tc.GetParameterValue("test"); r != v {
			t.Errorf("Expected %v, got %v", spew.Sdump(v), spew.Sdump(r))
		}

	}
}
