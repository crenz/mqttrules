// Wrap the script in a method, so that you can do "return false;" in case of an error or stop request
function publish()
{
        for (i = 1; i <= 1000; i++) {
          mqttspy.publish("trigger1", "{\"value\": " + i + "}", 0, false);          
        }
        // This means all OK, script has completed without any issues and as expected
        return true;
}

publish();
