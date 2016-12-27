package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-colorable"
	"flag"
	mqttrules "github.com/crenz/mqttrules"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(colorable.NewColorableStdout())
}

func main() {
	log.Infoln("Starting mqtt-rules")

	pBroker := flag.String("broker", "tcp://localhost:1883", "MQTT broker URI (e.g. tcp://localhost:1883)")
	pUsername := flag.String("username", "", "(optional) user name for MQTT broker access");
	pPassword := flag.String("password", "", "(optional) password for MQTT broker access");

	flag.Parse()

	if (mqttrules.Connect(*pBroker, *pUsername, *pPassword)) {
		mqttrules.Subscribe()
		for
		{
			mqttrules.Listen()
		}
		mqttrules.Disconnect()
	}
}
