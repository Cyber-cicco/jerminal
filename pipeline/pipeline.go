package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Cyber-cicco/jerminal/config"
	"github.com/Cyber-cicco/jerminal/utils"
	"github.com/google/uuid"
)

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	Agent         *config.Agent               `json:"agent"` // Agent executing the Pipeline
	AgentProvider                             // function executed at runtime to provide the Agent to the pipeline
	Name          string                      // human readable name of the pipeline
	mainDirectory string                      // Base directory of the pipeline
	directory     string                      // Working directory for the pipeline.
	Id            uuid.UUID                   `json:"id"` // UUID
	CloneFrom     *uuid.UUID                  `json:"parent,omitempty"`
	TimeRan       uint32                      `json:"time-ran"` // Number of time the pipeline ran
	events        []pipelineEvents            // components to be executed
	Inerror       bool                        `json:"in-error"` // Indicate if a fatal error has been encountered
	globalState   *config.GlobalStateProvider // L'état de l'application
	StartTime     time.Time                   `json:"start-time"`  // Début de la pipeline
	Diagnostic    *Diagnostic                 `json:"diagnostics"` // Infos about the current process. It can change based on what stage is getting executed.
	ElapsedTime   int64                       // Time it took to run the Pipeline

	// Copy of the config that should be initialized at start of
	// the pipeline so it keeps it's config even if there is a change during the execution
	Config *config.Config `json:"-"`
	Report *Report        `json:"-"` // Config that allows to choose a way of logging the results into a file
}

// Provides the agent for the pipeline
type AgentProvider func(p *Pipeline) *config.Agent

// Launches the events of the pipeline
//
// MUST BE CALLED IN A GOROUTINE BY THE SERVER
func (p *Pipeline) ExecutePipeline(ctx context.Context) error {
	var lastErr error
    p.Agent = p.AgentProvider(p)
	fmt.Printf("p.id: %v\n", p.Id)
	p.StartTime = time.Now()

	diag := NewDiag(fmt.Sprintf("%s", p.Name))

	p.Diagnostic = diag
	diag.NewDE(INFO, "Config was cloned")
	// Clean up work from the agent at end of pipeline
	defer func() {
		err := p.Agent.CleanUp()
		if err != nil {
			diag.NewDE(CRITICAL, fmt.Sprintf("Agent could not terminate properly because of error %v", err))
		}
		lastErr = err
		p.ElapsedTime = time.Now().UnixMilli() - p.StartTime.UnixMilli()
		diag.NewDE(INFO, fmt.Sprintf("Pipeline finished in %d ms", p.ElapsedTime))
		p.Report.Report(p)
	}()

	path, err := p.Agent.Initialize()

	if err != nil {
		fmt.Printf("err: %v\n", err)
		diag.NewDE(CRITICAL, fmt.Sprintf("Agent could not initialize because of error %v", err))
		return err
	}

	// Sets up the infos about the directory it will work in
	p.mainDirectory = path
	p.directory = p.mainDirectory

	pipePath := filepath.Join(p.globalState.PipelineDir, p.Id.String())

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

	diag.NewDE(INFO, "starting main loop")
	//Executes all the things from the pipeline
	for _, evt := range p.events {
		select {
		case <-ctx.Done():
			fmt.Printf("\"done\": %v\n", "done")
			diag.NewDE(WARN, "Pipeline got canceled before finishing")
			return ctx.Err()
		default:
			err := evt.ExecuteInPipeline(p, ctx)
			if err != nil {
				if evt.GetShouldStopIfError() {
					p.Inerror = true
					diag.NewDE(ERROR, fmt.Sprintf("got blocking error in executable %s : %v", evt.GetName(), err))
					break
				}
			}
		}
	}
	diag.NewDE(DEBUG, "End of execution")
	return lastErr
}

// ReportJson tells the pipeline to create a JSON file of
// Diagnostic after the execution of the pipeline, wether
// it fails or not
func (p *Pipeline) ReportJson() {
	p.Report.Types = append(p.Report.Types, JSON)
}

func (p *Pipeline) ReportHTML() {
	p.Report.Types = append(p.Report.Types, HTML)
}

func (p *Pipeline) ReportSQLITE() {
	p.Report.Types = append(p.Report.Types, SQLITE)
}

func (p *Pipeline) SetReportLogLevel(imp DEImp) {
	p.Report.LogLevel = imp
}

// Clone gives back a shallow copy of the Pipeline
//
// Pipelines share their executables and their agent.
//
// Executed pipelines also share the same ClonedFrom property,
// which corresponds to the ID they were cloned from
// (hence the name)
func (p *Pipeline) Clone() Pipeline {
	pipeline := *p
	pipeline.Id = uuid.New()
	pipeline.CloneFrom = &p.Id
	return pipeline
}

// SetPipeline initializes a new pipeline with the specified agent and components.
//
// It gets the current config of the app and gives back the Pipeline
func SetPipeline(name string, agent AgentProvider, events ...pipelineEvents) (*Pipeline, error) {
	s, err := config.GetState()
	if err != nil {
		return nil, err
	}
	return setPipelineWithState(name, agent, s, events...), nil
}

// setPipelineWithState gets a new pipeline with a config
//
// Only in testing should it be used by something else than SetPipeline
func setPipelineWithState(name string, agentProvider AgentProvider, config *config.GlobalStateProvider, events ...pipelineEvents) *Pipeline {
	p := Pipeline{
		Name:          name,
		Id:            uuid.New(),
        AgentProvider: agentProvider,
		mainDirectory: "",
		directory:     "",
		events:        events,
		Diagnostic:    &Diagnostic{},
		TimeRan:       0,
		globalState:   config,
		Report: &Report{
			Types:     []ReportType{},
			Directory: "./reports",
			LogLevel:  INFO,
		},
	}
	return &p
}
func (p *Pipeline) ResetDiag() {
	p.Diagnostic = p.Diagnostic.parent
}

// Agent retrieves an agent with the specified identifier.
func Agent(id string) AgentProvider {
	return func(p *Pipeline) *config.Agent {
		return p.globalState.GetAgent(id)
	}
}

// Returns the first agent available. If none is, returns
// the default agent
func AnyAgent() AgentProvider {
	return func(p *Pipeline) *config.Agent {
		return p.globalState.GetAnyAgent()
	}
}

// Returns the default agent even if busy
func DefaultAgent() AgentProvider {
	return func(p *Pipeline) *config.Agent {
		return p.globalState.DefaultAgent()
	}
}

func (p *Pipeline) GetId() string {
	return p.Id.String()
}
