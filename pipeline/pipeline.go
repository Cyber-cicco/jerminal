package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cyber-cicco/jerminal/state"
	"github.com/Cyber-cicco/jerminal/utils"
	"github.com/google/uuid"
)

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	Agent         *state.Agent           // Agent executing the Pipeline
	Name          string                 // human readable name of the pipeline
	mainDirectory string                 // Base directory of the pipeline
	directory     string                 // Working directory for the pipeline.
	id            uuid.UUID              // UUID
	timeRan       uint32                 // Number of time the pipeline ran
	events        []pipelineEvents       // components to be executed
	inerror       bool                   // Indicate if a fatal error has been encountered
	State         *state.ApplicationState // L'état de l'application
	Diagnostic    *Diagnostic            // Infos about de the process

	// Copy of the config that should be initialized at start of
	// the pipeline so it keeps it's state even if there is a change during the execution
	Config *state.Config
}

// Provides the agent for the pipeline
type AgentProvider func(p *Pipeline) *state.Agent

// Launches the events of the pipeline
//
// MUST BE CALLED IN A GOROUTINE BY THE SERVER
func (p *Pipeline) ExecutePipeline() error {
	var lastErr error

	diag := NewDiag(fmt.Sprintf("%s#%s", p.Name, p.id.String()))
	diag.NewDE(INFO, "starting main loop")

	p.Diagnostic = diag
    p.Config = p.State.CloneConfig()


    // Clean up work from the agent at end of pipeline
	defer func() {
		err := p.Agent.CleanUp()
		if err != nil {
            diag.NewDE(CRITICAL, fmt.Sprintf("agent could not terminate properly because of error %v", err))
        }
		lastErr = err
	}()

	path, err := p.Agent.Initialize()

	if err != nil {
        fmt.Printf("err: %v\n", err)
		diag.NewDE(CRITICAL, fmt.Sprintf("agent could not initialize because of error %v", err))
		return err
	}

    // Sets up the infos about the directory it will work in
	p.mainDirectory = path
	p.directory = p.mainDirectory

    pipePath := filepath.Join(p.State.PipelineDir, p.id.String())

    _, err = os.Stat(pipePath)

    // Create the directory for the pipeline if it does not yet exist
    if err != nil {
        err := os.MkdirAll(pipePath, os.ModePerm)
        if err != nil {
            return err
        }
    } else {
        err := utils.CopyDir(pipePath, p.mainDirectory)
        if err != nil {
            return err
        }
    }

    //Executes all the things from the pipeline
	for _, comp := range p.events {
		err := comp.ExecuteInPipeline(p)
		if err != nil {
			if comp.GetShouldStopIfError() {
				p.inerror = true
				diag.NewDE(ERROR, fmt.Sprintf("got blocking error in executable %s : %v", comp.GetName(), err))
				break
			}
		}
	}
	return lastErr
}

// SetPipeline initializes a new pipeline with the specified agent and components.
//
// It gets the current state of the app and gives back the Pipeline
func SetPipeline(name string, agent AgentProvider, events ...pipelineEvents) (*Pipeline, error) {
	s, err := state.GetState()
	if err != nil {
		return nil, err
	}
	return setPipelineWithState(name, agent, s, events...), nil
}

// setPipelineWithState gets a new pipeline with a state
//
// Only in testing should it be used by something else than SetPipeline
func setPipelineWithState(name string, agent AgentProvider, state *state.ApplicationState, events ...pipelineEvents) *Pipeline {
	p := Pipeline{
		Name:          name,
		id:            uuid.New(),
		mainDirectory: "",
		directory:     "",
		events:        events,
		Diagnostic:    &Diagnostic{},
		timeRan:       0,
		State:         state,
	}
	p.Agent = agent(&p)
	return &p
}

// Agent retrieves an agent with the specified identifier.
func Agent(id string) AgentProvider {
	return func(p *Pipeline) *state.Agent {
		return p.State.GetAgent(id)
	}
}

func AnyAgent() AgentProvider {
    return func(p *Pipeline) *state.Agent {
        return p.State.GetAnyAgent()
    }
}

func DefaultAgent() AgentProvider {
    return func(p *Pipeline) *state.Agent {
        return p.State.DefaultAgent()
    }
}
