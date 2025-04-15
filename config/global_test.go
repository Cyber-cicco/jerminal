package config

import (
	"os"
	"testing"

	"github.com/Cyber-cicco/jerminal/utils"
)

var _test_count int = 1

func TestAgentLifeCycle(t *testing.T) {
	var STATE GlobalStateProvider = GlobalStateProvider{
		Config: &Config{
			JerminalResourcePath: "../resources/jerminal.json",
		},
		agents: make(map[string]*Agent),
	}

	expectedPath :=  os.ExpandEnv("$HOME/.jerminal/agent/test")
	actualPath := ""

	STATE.UpdateConfig()
	a := STATE.GetAgent("test")
	actualPath, err := a.Initialize()
    utils.FatalError(err, t)
    utils.FatalExpectedActual(expectedPath, actualPath, t)

    err = a.CleanUp()
    utils.FatalError(err, t)

    if _test_count < 2 {
        _test_count++
        TestAgentLifeCycle(t)
    }

}
