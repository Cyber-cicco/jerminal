package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Configurable variables for the app
type Config struct {
	sync.RWMutex        // Can be read and written to by multiple routines, so we should lock it
	AgentDir     string `json:"agent-dir"`    // Source directory where agents do their work
	PipelineDir  string `json:"pipeline-dir"` // Source directory where pipelines cache the results of commands that should run once
}

var CONF Config

// InitConfig Creates the config object
//
// It should only be called by the server
func InitConfig() {
	CONF.Lock()
	defer CONF.Unlock()
	file, err := os.ReadFile("./resources/jerminal.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(file, &CONF); err != nil {
		panic(err)
	}
	CONF.setupEnv()
}

// setupEnv allows to read the env variables in the config file
func (c *Config) setupEnv() {
	c.AgentDir = os.ExpandEnv(c.AgentDir)
	c.PipelineDir = os.ExpandEnv(c.PipelineDir)
}
