[![Build Status](https://travis-ci.org/crenz/mqttrules.svg?branch=master)](https://travis-ci.org/crenz/mqttrules)
[![Go Report Card](https://goreportcard.com/badge/github.com/crenz/mqttrules)](https://goreportcard.com/report/github.com/crenz/mqttrules)
[![Coverage Status](https://coveralls.io/repos/github/crenz/mqttrules/badge.svg?branch=master)](https://coveralls.io/github/crenz/mqttrules?branch=master)
[![GoDoc](https://godoc.org/github.com/crenz/mqttrules?status.svg)](https://godoc.org/github.com/crenz/mqttrules)

# mqttrules

mqttrules executes rules that operate on MQTT messages.

## Usage

### Command-line

```
go get mqttrules
# without configuration file
mqttrules --broker tcp://localhost:1883 --username user --password pass
# with configuration file
mqttrules --config test/testConfig.json
```

### Docker image

```
docker run -v /var/mqttrules:/var/mqttrules crenz/mqttrules
```


## Rule definition

Each rule belongs to a _ruleset_, has a _name_, is either _triggered_ through
an incoming MQTT message with a certain topic or run according to a cron-like
_schedule_. If an (optional) _condition_ evaluates to true, one or several
_actions_ are performed. The only kind of action currently available is
sending out an MQTT message.

Here is an example. It refers to a _parameter_ called lights_kitchen_state.` This will be explained later.

```
{
        "trigger": "home/buttons/kitchen/status",
        "condition": "payload() > 0",
        "actions": [
          {
            "topic": "home/lights/kitchen/set",
            "payload": "{ \"on\" : ${!lights_kitchen_state}}",
            "qos": 2,
            "retain": false
          }
        ]
}
 ```

See [testConfig.json](../blob/master/test/testConfig.json)
for more examples, and for how to specify rules within the configuration file.

### Defining rules via MQTT messages

To define a rule using MQTT messages, send a message using the topic `rule/$RULESET/$RULENAME`.
As payload, send a JSON string in the format of the example above. In the
configuration file, you can also specify a prefix. With a prefix of `mqttrules/`,
the topic could be e.g. `mqttrules/rule/lights/kitchen_switch.

## Parameters

Parameters are values that can be used both as part of
_condition expressions_ as well as _payload expressions_. Parameters can have
an initial _value_, but they also can be defined based upon receiving messages
with a certain _topic_ and evaluating an _expression_ that defines the parameter value.

The following example defines a parameter that extracts the `on` value out
of MQTT messages with a JSON payload and the topic `home/lights/kitchen/status`.

```
{
      "value": "",
      "topic": "home/lights/kitchen/status",
      "expression": "payload(\"$.on\")"
}
```


### Defining parameters via MQTT messages

To define a parameter using MQTT messages, send a message using the
topic `param/$PARAMNAME`. As payload, send a JSON string in the format of the example above. In the
configuration file, you can also specify a prefix. With a prefix of `mqttrules/`,
the topic could be e.g. `mqttrules/param/lights_kitchen_state.`

**Note:** It is highly recommended to only use the characters `[A-Za-z0-9_]`
for the rule names, and to avoid especially the minus sign.

## Expressions

In mqttrules, expressions can make use of standard arithmetic expressions,
parameters and the special `payload()` function. Without a parameter,
`payload()` returns the complete payload of the incoming MQTT message
(that triggered the rule execution or parameter update). Alternatively,
you can specify a JSON path as parameter, e.g. `payload("$.state.on")`.

Condition expressions (used in rules) need to evaluate to a boolean value.
Parameter expressions (used to determine the value of a parameter) can evaluate
to any kind of value.

Expressions are evaluated using the [govaluate](https://github.com/Knetic/govaluate) package. For
more information, refer to that package's documentation.

## Kudos

mqttrules is made possible by leveraging some awesome Go packages:

* [cron](https://github.com/robfig/cron) - A cron spec parser and runner
* [govaluate](https://github.com/Knetic/govaluate) - Arbitrary expression evaluation for golang
* [jsonpath](https://github.com/oliveagle/jsonpath) - A golang implementation of JsonPath syntax.
* [gosweep|https://github.com/h12w/gosweep] - A shell script to do various checks on Go code.

## License

Copyright (C) 2016-2017 Christian Renz. Licensed under the MIT License (see LICENSE).


