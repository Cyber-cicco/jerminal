package config

import (
	"encoding/json"
	"os"
	"reflect"
	"sync"
)

// Configurable variables for the app
type Config struct {
	sync.RWMutex                                // Can be read and written to by multiple routines, so we should lock it
	AgentDir             string                 `json:"agent-dir"`    // Source directory where agents do their work
	PipelineDir          string                 `json:"pipeline-dir"` // Source directory where pipelines cache the results of commands that should run once
	ReportDir            string                 `json:"report-dir"`
	JerminalResourcePath string                 `json:"-"`
	AgentResourcePath    string                 `json:"-"`
	GithubWebhookSecret  string                 `json:"github-webhook-secret"` // pour l'authentification des webhooks github
	Secret               string                 `json:"secret"`
	UserParams           map[string]interface{} `json:"project"`
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

// setupEnv allows for the valorisation of env variables from the
// json file
func (c *Config) setupEnv() {
	expandStringFields(reflect.ValueOf(c).Elem())
}

func expandStringFields(v reflect.Value) {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		expandStringFields(v.Elem())
		return
	}

	switch v.Kind() {
	case reflect.String:
		if v.CanSet() {
			v.SetString(os.ExpandEnv(v.String()))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			expandStringFields(v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			expandStringFields(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			expandStringFields(v.MapIndex(key))
		}
	}
}
