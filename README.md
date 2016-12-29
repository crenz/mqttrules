[![Build Status](https://travis-ci.org/crenz/mqttrules.svg?branch=master)](https://travis-ci.org/crenz/mqttrules)

# mqttrules

mqttrules executes rules that operate on MQTT messages. The rules are defined by sending MQTT messages.

**NOTE: This project is currently in the design stage. There's no code yet.**

## 

Rules are defined as part of _rule sets_. Each rule set can be enabled or disabled. In addition, each rule can be enabled or disabled.

Rules can be activated in two ways

* through incoming MQTT messages (_triggers_)
* through a cron-like _schedule_

Once a rule is activated, one or several (optional) _conditions_ are checked. If the conditions are met, the specified _actions_ are executed. 

Possible actions are
* Send out one or more MQTT messages.

When sending out messages, a template mechanism can be used to customize the message and message content being sent out. 
In those templates, you can refer to pre-defined _parameters_ as well as to the JSON content of the trigger message.

### Rule examples

#### Use cases for testing viability of specification:

##### Lighting

1. Use a switch to switch off all lamps

2. Use an on/off switch to toggle a lamp

3. Use an on/off switch to toggle several lamps

4. Switch to a certain scene at 7pm every night. Switch to another scene at 7am.

5. Set all lights to the value passed in the payload of a message

##### Heating

1. Set all thermostats to $NIGHT_TEMP °C at night

2. Use a switch to toggle vacation mode. Vacation mode sets $DAY_TEMP to $DAY_TEMP_VACATION_LEVEL

3. Set a thermostat to a lower temperature when opening a certain window, set it back when closing it.

##### Sensors

1. Send an alert if humidity goes above a certain level

2. Reformat sensor values to a more agreeable format and send them out again

#### Examples

`rule/lighting/switch_on`

```json
{
	"trigger": "",
	"conditions": ["$lights-livingroom-status$ > 0"],
	"actions": [{
		"topic": "lights/livingroom/set",
		"payload": "{ \"value\" : 0}",
		"qos": 2,
		"retain": false
	}]
}




```

## Defining parameters

mqttrules treats condition and action payload expressions in a template-like fashion. You can set parameters that then can be used to customize these expressions.

The mqttrules parameters all live in the `param` topic space. You can set a parameter by sending a message with the topic `param/$PARAM_NAME`.

Parameters can be set either to a fixed value, or they can be tied to MQTT message payloads. In the latter case, mqttrules will subscribe to the respective MQTT topics and update the parameter value when new messages are received.

### Use cases for testing viability of specification

1. Set a parameter to a fixed value

2. Tie a parameter to a message payload

3. Tie a parameter to a JSON expression operating on a message payload

### Examples

Sent to the appropriate topic, the following payload will set a parameter to the fixed value `"42"`.

```json
{
    "value": "42"
}
```

Sent to the appropriate topic, the following payload will tie a parameter value to the payload of the topic `lighting/livingroom/status`:

`Topic: param/test-tie-payload`
```json
{
    "tie": "/lighting/livingroom/status"
}
```

Sent to the appropriate topic, the following payload will tie a parameter value to the payload of the topic `lighting/livingroom/status`. The payload will be interpreted as JSON, and the parameter will be set to whatever ".value" evaluates to:

`param/test-tie-json-payload`
```json
{
    "tie": "/lighting/livingroom/status",
    "json-path": ".value" 
}
```


## Components

* `cmd/mqttrules/mqttrules.go`: utility to run mqttrules

## License

Copyright (C) 2016 Christian Renz. Licensed under the MIT License (see LICENSE).


