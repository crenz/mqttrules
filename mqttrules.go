package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/crenz/mqttrules/agent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mattn/go-colorable"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(colorable.NewColorableStdout())
}

const defaultBroker string = "tcp://localhost:1883"

func getLogLevel(llString string) log.Level {
	level, error := log.ParseLevel(llString)
	if error != nil {
		log.Errorf("Unknown log level %s, using info", llString)
		return log.InfoLevel
	}
	return level
}

func main() {
	pBroker := flag.String("broker", "", "(optional) MQTT broker URI (e.g. tcp://localhost:1883)")
	pConfigFile := flag.String("config", "", "(optional) configuration file")
	pUsername := flag.String("username", "", "(optional) user name for MQTT broker access")
	pPassword := flag.String("password", "", "(optional) password for MQTT broker access")
	pLogLevel := flag.String("loglevel", "", "(optional) logging level (panic, fatal, error, warn, info, debug)")

	flag.Parse()

	var c *agent.ConfigFile
	var err error
	if len(*pConfigFile) > 0 {
		c, err = agent.ConfigFromFile(*pConfigFile)
		if err != nil {
			log.Errorf("Error reading config file %s: %v", *pConfigFile, err)
			return
		}
	} else {
		c = &agent.ConfigFile{}
	}

	// Arguments given via command line have precedence
	if len(*pBroker) > 0 {
		c.Config.Broker = *pBroker
	}
	if len(c.Config.Broker) == 0 {
		c.Config.Broker = defaultBroker
	}
	if len(*pUsername) > 0 {
		c.Config.Username = *pUsername
	}
	if len(*pPassword) > 0 {
		c.Config.Username = *pPassword
	}
	if len(*pLogLevel) > 0 {
		c.Config.Loglevel = *pLogLevel
	}

	log.SetLevel(getLogLevel(c.Config.Loglevel))

	log.Infoln("mqtt-rules connecting to broker", c.Config.Broker)
	opts := mqtt.NewClientOptions()
	opts.SetClientID(c.Config.ClientID)
	opts.AddBroker(c.Config.Broker)
	opts.SetUsername(c.Config.Username)
	opts.SetPassword(c.Config.Password)
	mqttClient := agent.NewPahoClient(opts)

	a := agent.New(mqttClient, c.Config.Prefix)

	if a.Connect() {
		a.InjectConfigFile(*c)
		a.Subscribe()

		for {
			a.Listen()
		}
		// will never be reached: c.Disconnect()
	}
}
