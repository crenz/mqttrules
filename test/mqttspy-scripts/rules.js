// Wrap the script in a method, so that you can do "return false;" in case of an error or stop request
function publish()
{
        mqttspy.publish("rule/test/trigger", "{\"trigger\": \"trigger1\", \"actions\": [{\"topic\": \"triggeredby\/trigger\", \"payload\": \"testing\"}]}", 1, true);
        mqttspy.publish("rule/test/schedule", "{\"schedule\": \"@every 10s\", \"actions\": [{\"topic\": \"triggeredby\/schedule\", \"payload\": \"testing\"}]}", 1, true);

        // This means all OK, script has completed without any issues and as expected
        return true;
}

publish();
