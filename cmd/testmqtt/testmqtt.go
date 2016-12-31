package main

import (
	"fmt"

	"github.com/eclipse/paho.mqtt.golang"
)

func init() {
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:9883")
	opts.SetClientID("test")
	opts.SetCleanSession(true)

	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for i := 0; i < 5; i++ {
		fmt.Println("---- doing publish ----")
		token := client.Publish("test", 1, false, fmt.Sprintf("%v", i))
		token.Wait()
	}
}
