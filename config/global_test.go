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

func TestResourceCreation(t *testing.T) {
    execPath, err := os.Getwd()
	homeDirEnv := os.Getenv("USERPROFILE")
	if homeDirEnv == "" {
		homeDirEnv = os.Getenv("HOME")
	}
    var conf Config
	conf.AgentDir = homeDirEnv + "/.jerminal/agent"
	conf.PipelineDir = homeDirEnv + "/.jerminal/pipeline"
	conf.ReportDir = execPath + "./reports"
    utils.FatalError(err, t)
    initializeApplicationResources(&conf, execPath, homeDirEnv)
}
