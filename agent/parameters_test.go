package agent

import (
	"testing"

	"github.com/crenz/mqttrules/test"
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
		testClient.SetParameter(c.key, c.value)
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

	testClient.SetParameter(key, `{
    "value": 42.2,
    "topic": "lighting/livingroom/status",
    "jsonPath": "$.value"
}`)
	resultFloat := testClient.GetParameterValue(key).(float64)
	if resultFloat != 42.2 {
		t.Errorf("GetParameterValue(%q) == %v, want %v", key, result, 42.2)
	}

	testClient.SetParameter(key, `{
    "value": 42,
    "topic": "lighting/livingroom/status",
    "jsonPath": "$.value"
}`)
	resultFloat = testClient.GetParameterValue(key).(float64)
	if resultFloat != 42 {
		t.Errorf("GetParameterValue(%q) == %v, want %v", key, result, 42)
	}
}

func TestReplaceParamsInString(t *testing.T) {
	testClient := New(test.NewClient(), "")

	testClient.SetParameter("test1", "value1")
	for _, c := range []struct {
		template, expected string
	}{
		{"", ""},
		{"$$", "$"},
		{"$test1$", "value1"},
		{"testing$test1$123", "testingvalue1123"},
		{"testing$test2$123", "testing123"},
	} {
		result := testClient.ReplaceParamsInString(c.template)
		if result != c.expected {
			t.Errorf("ReplaceParamsInString(%q); expected='%q', actual='%q'", c.template, c.expected, result)
		}
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

/*
func ExampleSetParameter() {
	//	c := NewClient()

	//	c.SetParameter("/rules/TTL", "60")
}

func ExampleGetParameterValue() {
	//	c := NewClient()

	//	c.GetParameterValue("/rules/TTL")

	//	result := GetParameter("/rules/TTL")

}
*/
