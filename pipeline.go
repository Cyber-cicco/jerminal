package jerminal

import (
	"fmt"
	"sync"
	"time"

	"github.com/Cyber-cicco/jerminal/state"
	"github.com/google/uuid"
)

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	Agent         *state.Agent           // Agent executing the Pipeline
	name          string                 // human readable name of the pipeline
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

type DEImp uint8

const (
	DEBUG = DEImp(iota)
	INFO
	WARN
	ERROR
	CRITICAL
)

// Informations about an element of the pipeline
type Diagnostic struct {
	label        string             // Name of the diagnostic
	identifier   uuid.UUID          // Unique identifier of the diagnostic
	date         time.Time          // Time the diagnostic was written
	inerror      bool               // Tells if the attached process should be considered in error
	events       []*DiagnosticEvent // Infos about what happened in the process
	sync.RWMutex                    // Can be used in goroutines so need to lock it
}

// NewDiag gets a Diagnostic with defaults
func NewDiag(name string) *Diagnostic {
	return &Diagnostic{
		label:      name,
		identifier: uuid.New(),
		date:       time.Now(),
		inerror:    false,
		events:     []*DiagnosticEvent{},
	}
}

// NewDE is a helper function to add an event to the diagnostic
func (d *Diagnostic) NewDE(importance DEImp, description string) {
	d.events = append(d.events, &DiagnosticEvent{
		importance:  importance,
		description: fmt.Sprintf("%v : %s", time.Now(), description),
	})
}

// Infos about an event
type DiagnosticEvent struct {
	importance  DEImp  // Importance of the event
	description string // Description of the event
}

// Launches the events of the pipeline
//
// MUST BE CALLED IN A GOROUTINE BY THE SERVER
func (p *Pipeline) ExecutePipeline() error {
	var lastErr error

	diag := NewDiag(fmt.Sprintf("%s#%s", p.name, p.id.String()))
	diag.NewDE(INFO, "starting main loop")

	p.Diagnostic = diag
    p.Config = p.State.CloneConfig()

	defer func() {
		err := p.Agent.CleanUp()
		diag.NewDE(CRITICAL, fmt.Sprintf("agent could not terminate properly because of error %v", err))
		lastErr = err
	}()

	path, err := p.Agent.Initialize()

	if err != nil {
		diag.NewDE(CRITICAL, fmt.Sprintf("agent could not initialize because of error %v", err))
		return err
	}

	p.mainDirectory = path
	p.directory = p.mainDirectory

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
