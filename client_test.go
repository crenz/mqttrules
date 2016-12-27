package mqttrules

import (
	"testing"
	"fmt"
)

func TestSendParam(t *testing.T) {
	messages = make(chan [2]string)

	for _, c := range []struct {
		key, value string
	}{
		{"param", "value"},
		{"param", ""},
		{"", ""},
	} {
		go func() { messages <- [2]string{fmt.Sprintf("mr/param/%s", c.key), c.value} }()
		Listen()
		result := GetParameter(c.key)
		if result != c.value {
			t.Errorf("GetParameter(%q) == %q, want %q", c.key, result, c.value)
		}
	}


}
