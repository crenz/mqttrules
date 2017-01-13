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
	testClient.setPrefix("mr/")

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

func TestAgent_HandleMessage(t *testing.T) {
	for _, prefix := range []string{"", "prefix", "prefix/"} {
		mqttClient := test.NewClient()
		testClient := New(mqttClient, prefix)

		testClient.HandleMessage(fmt.Sprintf("%sparam/test1", prefix), []byte(`{"value": 42}`))
		if r := testClient.GetParameterValue("test1"); r != 42.0 {
			t.Errorf("[Prefix = '%s'], Parameter value is %v, should be 42", prefix, spew.Sdump(r))
		}
		testClient.HandleMessage(fmt.Sprintf("param/test2", prefix), []byte(`{"value": 42}`))
		if r := testClient.GetParameterValue("test2"); r == 42.0 && len(prefix) > 0 {
			t.Errorf("[Prefix = '%s'], Parameter should not have been set", prefix)
		}
	}
}
