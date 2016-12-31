package mqttrules

import (
	"testing"
	"fmt"
)

func TestSendParam(t *testing.T) {
	// Don't use NewClient to enable access to private field messages
	testClient := &client{}
	testClient.initialize()
	testClient.SetPrefix("mr/")

	testClient.messages = make(chan [2]string)

	for _, c := range []struct {
		key, value string
	}{
		{"param", "value"},
		{"param", ""},
		{"", ""},
	} {
		go func() { testClient.messages <- [2]string{fmt.Sprintf("mr/param/%s", c.key), c.value} }()
		testClient.Listen()
		result := testClient.GetParameterValue(c.key)
		if result != c.value {
			t.Errorf("GetParameter(%q) == %q, want %q", c.key, result, c.value)
		}
	}


}
