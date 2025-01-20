package state

import (
	"errors"
	"os"
	"path"
	"sync"
)

// ApplicationState represents global state of the application
type ApplicationState struct {
	sync.RWMutex
	Config
	agents map[string]*Agent // map of identifiers to their agent
}

// Agent represents a process that executes a pipeline in its personal directory
// and cleans it up afterward. The Identifier uniquely identifies the agent.
type Agent struct {
	sync.Mutex
	BusySig    *sync.Cond // Signal informing if the Agent is busy
	identifier string     // unique string representing an Agent
	Busy       bool       // true if the agent is executing a pipeline
}

// STATE represents a mutable structure of application state
var STATE ApplicationState = ApplicationState{
	Config: Config{},
	agents: make(map[string]*Agent),
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
	path := path.Join(STATE.AgentDir, a.identifier)
	_, err := os.Stat(path)
	if err == nil {
		return "", errors.New("directory should not exist, agent has not cleaned up his directory from previous job")
	}
	return path, os.Mkdir(path, os.ModePerm)
}

// CleanUp sends a signal to tell it is not busy anymore,
// and cleans up it's directory
func (a *Agent) CleanUp() error {

    defer a.Unlock()
	a.Lock()

	path := path.Join(STATE.AgentDir, a.identifier)
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
			identifier: id,
			Busy:       false,
		}
		ag.BusySig = sync.NewCond(&ag.Mutex)
		s.agents[id] = ag
	}
	return ag
}
