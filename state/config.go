package state

import (
	"encoding/json"
	"os"
	"sync"
)

// Configurable variables for the app
type Config struct {
	sync.RWMutex                // Can be read and written to by multiple routines, so we should lock it
	AgentDir             string `json:"agent-dir"`    // Source directory where agents do their work
	PipelineDir          string `json:"pipeline-dir"` // Source directory where pipelines cache the results of commands that should run once
	JerminalResourcePath string
	AgentResourcePath    string
}

// UpdateConfig Creates the config object
//
// It should only be called by the server
func (c *Config) UpdateConfig() error {
	c.Lock()
	defer c.Unlock()
	file, err := os.ReadFile(c.JerminalResourcePath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(file, c); err != nil {
		return err
	}
	c.setupEnv()
	return nil
}

// setupEnv allows to read the env variables in the config file
func (c *Config) setupEnv() {
	c.AgentDir = os.ExpandEnv(c.AgentDir)
	c.PipelineDir = os.ExpandEnv(c.PipelineDir)
}
