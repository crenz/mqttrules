package agent

import (
	"fmt"
	"testing"

	"strings"

	"github.com/crenz/mqttrules/test"
	"github.com/davecgh/go-spew/spew"
)

func TestClient_ConnectSubscribeDisconnect(t *testing.T) {
	mqttClient := test.NewClient()
	testClient := New(mqttClient, "")

	if !testClient.Connect() {
		t.Error("Failed to connect")
	}
	if !testClient.Subscribe() {
		t.Error("Failed to subscribe")
	}
	testClient.Disconnect()

}

func TestClient_Publish(t *testing.T) {
	mqttClient := test.NewClient()
	testClient := New(mqttClient, "")

	for _, c := range []struct {
		topic    string
		qos      byte
		retained bool
		payload  string
	}{
		{"", 0, false, ""},
	} {
		testClient.Publish(c.topic, c.qos, c.retained, c.payload)
		message := mqttClient.LastMessage()
		if strings.Compare(c.topic, message.Topic) != 0 || c.qos != message.QoS || c.retained != message.Retained {
			t.Errorf("Message was not published properly")
			spew.Dump(map[string]interface{}{"Test data": c, "Result": message})
		}
	}

}

func TestSendParam(t *testing.T) {
	// Don't use NewClient to enable access to private field messages
	testClient := &agent{}
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
