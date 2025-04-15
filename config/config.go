package config

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
	ReportDir            string `json:"report-dir"`
	JerminalResourcePath string
	AgentResourcePath    string
	GithubWebhookSecret  string `json:"github-webhook-secret"` // pour l'authentification des webhooks github
	Secret               string `json:"secret"`
}

type Project struct {
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
	c.ReportDir = os.ExpandEnv(c.ReportDir)
	c.Secret = os.ExpandEnv(c.Secret)
	c.GithubWebhookSecret = os.ExpandEnv(c.GithubWebhookSecret)
}
