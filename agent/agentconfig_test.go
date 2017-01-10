package agent

import (
	"os"
	"testing"
)

func TestConfigFromFile(t *testing.T) {
	path := "../test/testConfig.json"

	_, err := ConfigFromFile(path)
	if err != nil {
		wd, _ := os.Getwd()
		t.Errorf("Failed to read config file %s: %v (working dir: %s)", path, err, wd)
	}

}
