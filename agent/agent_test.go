package agent

import (
	"fmt"
	"testing"

	"strings"

	"os"

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

func Test_TestConfigJSON(t *testing.T) {
	path := "../test/testConfig.json"

	c, err := ConfigFromFile(path)
	if err != nil {
		wd, _ := os.Getwd()
		t.Errorf("Failed to read config file %s: %v (working dir: %s)", path, err, wd)
		return
	}

	mqttClient := test.NewClient()
	a := New(mqttClient, "")
	a.Connect()
	a.InjectConfigFile(*c)

	for p := range c.Parameters {
		if v := a.GetParameterValue(p); c.Parameters[p].Value != v {
			t.Errorf("Param %s: expected %v, got %v", p, c.Parameters[p].Value, v)
		}
	}

	for ruleset := range c.Rules {
		for rule := range c.Rules[ruleset] {
			if r := a.GetRule(ruleset, rule); r == nil {
				t.Errorf("Rule %s/%s was not created", ruleset, rule)
			}
		}
	}

	a.Subscribe()

	if v := a.GetParameterValue("lights_kitchen_state"); v != 0.0 {
		t.Errorf("Param lights_kitchen_state: expected initial value %v, got %v", spew.Sdump(0.0), spew.Sdump(v))
	}
	a.HandleMessage("home/lights/kitchen/status", []byte(` { "on": 1 } `))
	if v := a.GetParameterValue("lights_kitchen_state"); v != 1.0 {
		t.Errorf("Param lights_kitchen_state: expected updated value %v, got %v", spew.Sdump(1.0), spew.Sdump(v))
	}

	a.HandleMessage("home/buttons/kitchen/status", []byte("1"))
	message := mqttClient.LastMessage()

	if strings.Compare("home/lights/kitchen/set", message.Topic) != 0 {
		t.Errorf("Received unexpected message with topic %s", message.Topic)
		return
	}
	a.HandleMessage("home/lights/kitchen/status", []byte(message.Payload.(string)))
	if v := a.GetParameterValue("lights_kitchen_state"); v != 0.0 {
		t.Errorf("Param lights_kitchen_state: expected updated value %v, got %v", spew.Sdump(0.0), spew.Sdump(v))
		t.Errorf(spew.Sdump(message))
	}

	a.Disconnect()
}
