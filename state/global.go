package state

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
)

// ApplicationState represents global state of the application
type ApplicationState struct {
	sync.RWMutex
	*Config
	agents map[string]*Agent // map of identifiers to their agent
}

// Agent represents a process that executes a pipeline in its personal directory
// and cleans it up afterward. The Identifier uniquely identifies the agent.
type Agent struct {
	sync.Mutex
	BusySig    *sync.Cond        // Signal informing if the Agent is busy
	Identifier string            // unique string representing an Agent
	Busy       bool              // true if the agent is executing a pipeline
	State      *ApplicationState // The application state
}

var (
	state *ApplicationState
	once  sync.Once
)

// Creates an object containing infos about the application process
//
// SHOULD ONLY BE CALLED ONCE
func initializeApplicationState(conf *Config) {
	state = &ApplicationState{
		agents: make(map[string]*Agent),
		Config: conf,
	}
}

// Gets the current state of the application with a default config. 
// Updates the state of the config, so if you change the config file
// between getting the previous state and this one, it will run with
// the updated config
//
// Mutually exclusive with GetStateCustomConf
//
// Should be used by default
func GetState() (*ApplicationState, error) {
	once.Do(func() {
        fmt.Printf("\"in once\": %v\n", "in once")
		conf := &Config{
            JerminalResourcePath: "./resources/jerminal.json",
        }
		initializeApplicationState(conf)
	})
	return state, state.UpdateConfig()
}

// Gets the current state of the application with a custom config. 
//
// Mutually exclusive with GetState
//
// Should be used for tests
func GetStateCustomConf(conf *Config) *ApplicationState {
	once.Do(func() {
        fmt.Printf("\"in once custom\": %v\n", "in once")
        initializeApplicationState(conf)
    })
	return state
}

// Initialize first waits until the agent has finished his
// previous work, then creates the directory it will work in
func (a *Agent) Initialize() (string, error) {
	// Wait until the agent is no longer busy
	a.Lock()
	for a.Busy {
		a.BusySig.Wait()
	}
	a.Busy = true
	a.Unlock()

	// Create the agent directory
	path := path.Join(a.State.AgentDir, a.Identifier)
	infos, err := os.Stat(path)
	if err == nil {
		return "", errors.New(fmt.Sprintf("directory should not exist, agent has not cleaned up his directory from previous job. %s", infos.Name()))
	}
	return path, os.Mkdir(path, os.ModePerm)
}

// CleanUp sends a signal to tell it is not busy anymore,
// and cleans up it's directory
func (a *Agent) CleanUp() error {

	defer a.Unlock()
	a.Lock()

	path := path.Join(a.State.AgentDir, a.Identifier)
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	a.Busy = false
	a.BusySig.Signal()

	return nil
}

// GetAgent returns an agent from the map of agents
//
// Creates it does not exist yet
func (s *ApplicationState) GetAgent(id string) *Agent {
	ag, ok := s.agents[id]
	if !ok {
		ag = &Agent{
			Identifier: id,
			Busy:       false,
			State:      s,
		}
		ag.BusySig = sync.NewCond(&ag.Mutex)
		s.agents[id] = ag
	}
	return ag
}

func (s *ApplicationState) CloneConfig() *Config {
    s.Config.Lock()
    defer s.Config.Unlock()
    conf := Config{
    	AgentDir:             s.AgentDir,
    	PipelineDir:          s.PipelineDir,
    	JerminalResourcePath: s.JerminalResourcePath,
    }
    return &conf
}
