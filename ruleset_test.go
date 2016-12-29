package mqttrules

import "testing"

func TestSetParameter(t *testing.T) {
	testClient := NewClient()

	for _, c := range []struct {
		key, value string
	}{
		{"/rules/TTL", "Testing"},
		{"/rules/TTL", ""},
		{"", ""},
	} {
		testClient.SetParameter(c.key, c.value)
		result := testClient.GetParameter(c.key)
		if result != c.value {
			t.Errorf("GetParameter(%q) == %q, want %q", c.key, result, c.value)
		}
	}
}

func TestReplaceParamsInString(t *testing.T) {
	testClient := NewClient()

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

func ExampleSetParameter() {
	c := NewClient()

	c.SetParameter("/rules/TTL", "60")
}

func ExampleGetParameter() {
	c := NewClient()

	c.GetParameter("/rules/TTL")

	//	result := GetParameter("/rules/TTL")

}