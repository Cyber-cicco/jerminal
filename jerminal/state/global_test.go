package state

import (
	"os"
	"testing"
)

var _test_count int = 1

func TestAgentLifeCycle(t *testing.T) {
	var STATE ApplicationState = ApplicationState{
		Config: &Config{
			jerminalResourcePath: "../../resources/jerminal.json",
		},
		agents: make(map[string]*Agent),
	}

	expectedPath :=  os.ExpandEnv("$HOME/.jerminal/agent/test")
	actualPath := ""

	STATE.UpdateConfig()
	a := STATE.GetAgent("test")
	actualPath, err := a.Initialize()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if expectedPath != actualPath {
		t.Fatalf("Expected %s, got %s", expectedPath, actualPath)
	}

    err = a.CleanUp()

    if err != nil {
        t.Fatalf("Expect no error, got %v", err)
    }

    if _test_count < 2 {
        _test_count++
        TestAgentLifeCycle(t)
    }

}
