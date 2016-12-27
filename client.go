package mqttrules

import (
	log "github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"regexp"
	"fmt"
)

//import "github.com/robfig/cron"

//TODO Refactor into a instanciable class

var client mqtt.Client
var messages chan[2]string
var prefix string = "mr/"

/* public functions */

func Connect(broker string, username string, password string) bool {
	log.Infoln("Connecting to MQTT broker", broker)

	opts := mqtt.NewClientOptions()
	opts.SetClientID("mqtt-rules/0.1")
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)

	messages = make(chan [2]string)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		messages <- [2]string{msg.Topic(), string(msg.Payload())}
	})


	client = mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if (token.Error() != nil) {
		log.Errorln(token.Error())
		return false
	}

	return true
}

func Subscribe() bool {
	if token := client.Subscribe("#", byte(1), nil); token.Wait() && token.Error() != nil {
		log.Errorln(token.Error())
		return false
	}
	log.Infoln("Subscribed successfully")
	return true
}

func Listen() {
	incoming := <- messages
	//log.Infof("Received [%s] %s\n", incoming[0], incoming[1])
	r := regexp.MustCompile(fmt.Sprintf("^%sparam/([^/]+)", prefix))
	if res := r.FindStringSubmatch(incoming[0]); res != nil {
		log.Infof("Submatch [%s] in [%s] %s\n", res[1], incoming[0], incoming[1])
		SetParameter(res[1], incoming[1])
	}
}

func Disconnect() {
	log.Infoln("Disconnecting from MQTT broker")
	client.Disconnect(250)
}
