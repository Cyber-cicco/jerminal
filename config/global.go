package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
)

// GlobalStateProvider represents global config of the application
type GlobalStateProvider struct {
	*Config
	agents map[string]*Agent // map of identifiers to their agent
}

// Agent represents a process that executes a pipeline in its personal directory
// and cleans it up afterward. The Identifier uniquely identifies the agent.
type Agent struct {
	sync.Mutex
	BusySig    *sync.Cond           `json:"-"`          // Signal informing if the Agent is busy
	Identifier string               `json:"identifier"` // unique string representing an Agent
	Busy       bool                 `json:"-"`          // true if the agent is executing a pipeline
	State      *GlobalStateProvider `json:"-"`          // The application config
}

var (
	config *GlobalStateProvider
	once   sync.Once
)

const DEFAULT_AGENT = "6524a5fc-0772-4684-82d7-6900c444162b"

// Creates an object containing infos about the application process
//
// SHOULD ONLY BE CALLED ONCE
func initializeApplicationState(conf *Config) error {
	agents := []*Agent{}
	config = &GlobalStateProvider{
		Config: conf,
	}

	// Getting the agents from the config file
	file, err := os.ReadFile(conf.AgentResourcePath)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return errors.New("Process should have a ./resources/agents.json file in order to work. Check the docs to set it up")
	}

	if err = json.Unmarshal(file, &agents); err != nil {
		return err
	}

	//Creating the map of agents
	agentMap := make(map[string]*Agent)

	defaultAgent := &Agent{
		Identifier: DEFAULT_AGENT,
		Busy:       false,
		State:      config,
	}
	defaultAgent.BusySig = sync.NewCond(&defaultAgent.Mutex)
	agentMap[DEFAULT_AGENT] = defaultAgent

	for _, agent := range agents {
		newAgent := &Agent{
			Identifier: agent.Identifier,
			Busy:       false,
			State:      config,
		}
		newAgent.BusySig = sync.NewCond(&newAgent.Mutex)
		agentMap[newAgent.Identifier] = newAgent
	}

	config.agents = agentMap

	return nil
}

// Gets the current config of the application with a default config.
// Updates the config of the config, so if you change the config file
// between getting the previous config and this one, it will run with
// the updated config
//
// # Mutually exclusive with GetStateCustomConf
//
// Should be used by default
func GetState() (*GlobalStateProvider, error) {
	var err error
    if config == nil {
        err = onceStateCreator()
    }
	if err != nil {
		return nil, err
	}
	return config, config.UpdateConfig()
}

// onceStateCreator initializes the application state the first time
// GetState is called
//
// SHOULD ONLY BE CALLED ONCE BY GetState
func onceStateCreator() error {
	execPath, err := os.Getwd()
	if err != nil {
		return err
	}
	homeDirEnv := os.Getenv("USERPROFILE")
	if homeDirEnv == "" {
		homeDirEnv = os.Getenv("HOME")
	}
	conf := &Config{
		JerminalResourcePath: execPath + "/resources/jerminal.json",
		AgentResourcePath:    execPath + "/resources/agents.json",
        AgentDir: homeDirEnv + "/agent",
        PipelineDir: homeDirEnv + "/pipeline",
	}
	if _, err := os.Stat(conf.JerminalResourcePath); err != nil {
		initializeApplicationResources(conf, execPath, homeDirEnv)
	}

	if _, err := os.Stat(conf.AgentDir); err != nil {
        fmt.Printf("err: %v\n", err)
		fmt.Println("Jerminal env was not set up, creating the necessary directories...")
		err := os.MkdirAll(conf.AgentDir, 0644)
		if err != nil {
			fmt.Println("Encountered error while creating jerminal agent environnement. Terminating.", err)
            fmt.Printf("conf.AgentDir: %v\n", conf.AgentDir)
			os.Exit(1)
		}
		err = os.MkdirAll(conf.PipelineDir, 0644)
		if err != nil {
			fmt.Println("Encountered error while creating jerminal pipeline environnement. Terminating.")
			os.Exit(1)
		}
	}

	return initializeApplicationState(conf)

}

// initializeApplicationResources creates the necessary files and directories at
// the root of the project
func initializeApplicationResources(conf *Config, execPath, homeDirEnv string) {

	fmt.Println("Your project is not yet set up !")
	fmt.Print("Please enter a secret pass phrase for jerminal (prefix it with $ if you want it to be an env variable) : ")
	reader := bufio.NewReader(os.Stdin)
	var input string
	var err error
	for input == "" {
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Encountered error while reading input, terminating\n", err)
			os.Exit(1)
		}
		if input == "" {
			fmt.Println("The secret passphrase cannot be empty.")
			fmt.Print("Please enter again (prefix it with $ if you want it to be an env variable) : ")
		}
	}
	conf.AgentDir = homeDirEnv + "/.jerminal/agent"
	conf.PipelineDir = homeDirEnv + "/.jerminal/pipeline"
	conf.ReportDir = execPath + "./reports"
	conf.Secret = input

	if _, err := os.Stat(execPath + "/resources"); err != nil {
		os.Mkdir(execPath+"/resources", 0755)
		fmt.Println("Created a resources directory in the project root. This is where your project configuration files will be stored.")
	}
	confBytes, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(conf.JerminalResourcePath, confBytes, 0644)
	if err != nil {
		fmt.Println("Encountered error when creating the jerminal resource file. Terminating.")
		os.Exit(1)
	}
	fmt.Println("Created jerminal resource path.")
	agents := []Agent{
		{
			Identifier: "default",
		},
	}
	agentBytes, err := json.Marshal(agents)
	if err != nil {
		fmt.Println("Encountered error when creating the agents resource file. Terminating.")
		os.Exit(1)
	}
	err = os.WriteFile(conf.AgentResourcePath, agentBytes, 0644)
	if err != nil {
		fmt.Println("Encountered error when creating the agent resource file. Terminating.")
	}
	fmt.Println("Created agent resource path with a default agent. Add more agents if you want to execute tasks in parallel.")
	conf.Secret = os.ExpandEnv(conf.Secret)

}

// Gets the current config of the application with a custom config.
//
// # Mutually exclusive with GetState
//
// Should be used for tests
func GetStateCustomConf(conf *Config) *GlobalStateProvider {
	initializeApplicationState(conf)
	return config
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
	fmt.Println("Cleaning up")

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
func (s *GlobalStateProvider) GetAgent(id string) *Agent {
	s.Lock()
	defer s.Unlock()

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

// Gets the first non busy existing agent
//
// If every agent is busy, gets the default agent
func (s *GlobalStateProvider) GetAnyAgent() *Agent {

	s.Lock()
	defer s.Unlock()

	for _, agent := range s.agents {
		agent.Lock()
		available := !agent.Busy
		agent.Unlock()
		if available {
			return agent
		}
	}
	return s.agents[DEFAULT_AGENT]
}

// Gets the default agent back
func (s *GlobalStateProvider) DefaultAgent() *Agent {
	s.Lock()
	defer s.Unlock()
	return s.agents[DEFAULT_AGENT]
}

// Allows for retreival of configuration at an instant
// allowing for the pipeline to stay coherent even if a change
// to the config is made during it's runtime
func (s *GlobalStateProvider) CloneConfig() *Config {
	s.Lock()
	defer s.Unlock()
	s.Config.Lock()
	defer s.Config.Unlock()
	conf := Config{
		RWMutex:              sync.RWMutex{},
		AgentDir:             s.AgentDir,
		PipelineDir:          s.PipelineDir,
		ReportDir:            s.ReportDir,
		JerminalResourcePath: s.JerminalResourcePath,
		AgentResourcePath:    s.AgentResourcePath,
		GithubWebhookSecret:  s.GithubWebhookSecret,
		Secret:               s.Secret,
		UserParams:           s.UserParams,
	}
	return &conf
}
