package agent

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
)

func TestPahoClient(t *testing.T) {
	o := mqtt.ClientOptions{}
	o.AddBroker("pure-fantasy")
	c := NewPahoClient(&o)
	if c == nil {
		t.Error("Failed to create new Paho client")
	}
	if c.IsConnected() == true {
		t.Error("IsConnected returns wrong status")
	}
}
