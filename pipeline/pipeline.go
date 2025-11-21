package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Cyber-cicco/jerminal/config"
	"github.com/Cyber-cicco/jerminal/utils"
	"github.com/google/uuid"
)

type Key string

type PipelineParams struct {
	sync.Mutex
	params map[Key]interface{}
}

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	PipelineParams
	Agent         *config.Agent               `json:"agent"` // Agent executing the Pipeline
	agentProvider AgentProvider               // function executed at runtime to provide the Agent to the pipeline
	Name          string                      `json:"name"` // human readable name of the pipeline
	mainDirectory string                      // Base directory of the pipeline
	directory     string                      // Working directory for the pipeline.
	Id            uuid.UUID                   `json:"id"` // UUID
	CloneFrom     *uuid.UUID                  `json:"parent,omitempty"`
	TimeRan       uint32                      `json:"time-ran"` // Number of time the pipeline ran
	pipelineDir   string                      // Directory to cache things for subsequent runs of the pipeline
	events        []pipelineEvents            // components to be executed
	Inerror       bool                        `json:"in-error"` // Indicate if a fatal error has been encountered
	globalState   *config.GlobalStateProvider // L'état de l'application
	StartTime     time.Time                   `json:"start-time"` // Début de la pipeline
	EndTime       time.Time                   `json:"end-time"` // Fin de la pipeline
	Diagnostic    *Diagnostic                 `json:"diagnostics"`  // Infos about the current process. It can change based on what stage is getting executed.
	ElapsedTime   int64                       `json:"elapsed-time"` // Time it took to run the Pipeline

	// Copy of the config that should be initialized at start of
	// the pipeline so it keeps it's config even if there is a change during the execution
	Config *config.Config `json:"-"`
	Report *Report        `json:"report-type"` // Config that allows to choose a way of logging the results into a file
}

// Provides the agent for the pipeline
type AgentProvider func(p *Pipeline) *config.Agent

// Launches the events of the pipeline
//
// MUST BE CALLED IN A GOROUTINE BY THE SERVER
func (p *Pipeline) ExecutePipeline(ctx context.Context) error {
	var lastErr error
	p.Agent = p.agentProvider(p)
	p.StartTime = time.Now()

	diag := NewDiag(fmt.Sprintf("%s", p.Name))

	p.Diagnostic = diag
	// Clean up work from the agent at end of pipeline
	defer func() {
		err := p.Agent.CleanUp()
		if err != nil {
			diag.LogEvent(CRITICAL, fmt.Sprintf("Agent could not terminate properly because of error %v", err))
		}
		lastErr = err
        p.EndTime = time.Now()
		p.ElapsedTime = p.EndTime.UnixMilli() - p.StartTime.UnixMilli()
		diag.LogEvent(INFO, fmt.Sprintf("Pipeline finished in %d ms", p.ElapsedTime))
		if !p.Inerror {
			p.RanSuccessfully()
		}
		p.Report.Report(p)
	}()

	path, err := p.Agent.Initialize()

	if err != nil {
		fmt.Printf("err: %v\n", err)
		diag.LogEvent(CRITICAL, fmt.Sprintf("Agent could not initialize because of error %v", err))
		return err
	}

	// Sets up the infos about the directory it will work in
	p.mainDirectory = path
	p.directory = p.mainDirectory

	_, err = os.Stat(p.pipelineDir)

	// Create the directory for the pipeline if it does not yet exist
	if err != nil {
		err := os.MkdirAll(p.pipelineDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		err := utils.CopyDir(p.pipelineDir, p.mainDirectory)
		if err != nil {
			return err
		}
	}

	diag.LogEvent(INFO, "starting main loop")
	//Executes all the things from the pipeline
	for _, evt := range p.events {
		var stageErr error
		select {
		case <-ctx.Done():
			diag.LogEvent(WARN, "Pipeline got canceled before finishing")
			return ctx.Err()
		default:
			err := evt.ExecuteInPipeline(p, ctx)
			if err != nil {
				if evt.GetShouldStopIfError() {
					stageErr = err
					p.Inerror = true
					diag.LogEvent(ERROR, fmt.Sprintf("got blocking error in executable %s : %v", evt.GetName(), err))
				}
			}
		}
		if stageErr != nil {
			break
		}
	}

	diag.LogEvent(DEBUG, "End of execution")
	return lastErr
}

func (p *Pipeline) RanSuccessfully() {
	parent, ok := GetStore().GlobalPipelines[p.Name]
	if !ok {
		p.Diagnostic.LogEvent(WARN, "Parent must exist. If it was not ran from a test, it was definitly a problem")
		return
	}
	parent.TimeRan++
}

// ReportJson tells the pipeline to create a JSON file of
// Diagnostic after the execution of the pipeline, wether
// it fails or not
func (p *Pipeline) ReportJson() {
	p.Report.Types = append(p.Report.Types, JSON)
}

// ReportHTML tells the pipeline to store the content of
// the reports as HTML files using tailwind 3 for styling
// and vanilla JS for client side interactivity
//
// WARN : NOT YET IMPLEMENTED
func (p *Pipeline) ReportHTML() {
	p.Report.Types = append(p.Report.Types, HTML)
}

// ReportSQLITE tells the pipeline to store the content
// of the reports in an SQLite database
//
// WARN : NOT YET IMPLEMENTED
func (p *Pipeline) ReportSQLITE() {
	p.Report.Types = append(p.Report.Types, SQLITE)
}

func (p *Pipeline) SetReportLogLevel(imp EImportance) {
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
		Name:           name,
		Id:             uuid.New(),
		agentProvider:  agentProvider,
		mainDirectory:  "",
		directory:      "",
		events:         events,
		Diagnostic:     &Diagnostic{},
		TimeRan:        0,
		globalState:    config,
		Config:         config.Config,
		PipelineParams: PipelineParams{params: map[Key]interface{}{}},
		Report: &Report{
			Types:    []ReportType{},
			LogLevel: INFO,
		},
	}
	p.pipelineDir = filepath.Join(p.globalState.PipelineDir, p.Name)
	return &p
}
func (p *Pipeline) ResetDiag() {
	p.Diagnostic = p.Diagnostic.parent
}

type ResourceKey string

func (p *Pipeline) GetResource(param ResourceKey) (interface{}, bool) {
	res, ok := p.Config.UserParams[string(param)]
	return res, ok
}

func (p *Pipeline) MustGetResource(param ResourceKey) interface{} {
	return p.Config.UserParams[string(param)]
}

func (p *PipelineParams) Get(param Key) (interface{}, error) {
	p.Lock()
	defer p.Unlock()
	res, ok := p.params[param]
	if !ok {
		return nil, fmt.Errorf("param %s does not exist", param)
	}
	return res, nil
}

func (p *PipelineParams) MustGet(param Key) interface{} {
	p.Lock()
	defer p.Unlock()
	res, ok := p.params[param]
	if !ok {
		panic(fmt.Sprintf("param %s does not exist", param))
	}
	return res
}

func (p *PipelineParams) Put(key Key, val interface{}) {
	p.Lock()
	defer p.Unlock()
	p.params[key] = val
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
