package jerminal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	agent
	name          string           // human readable name of the pipeline
	MainDirectory string           // Base directory of the pipeline
	Directory     string           // Working directory for the pipeline.
	events        []pipelineEvents // components to be executed
	Diagnostics   []*Diagnostic
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
	label      string             // Name of the diagnostic
	identifier string             // Unique identifier of the diagnostic
	date       time.Time          // Time the diagnostic was written
	inerror    bool               // Tells if the attached process should be considered in error
	events     []*DiagnosticEvent // Infos about what happened in the process
}

// NewDiag gets a Diagnostic with defaults
func NewDiag(name string) *Diagnostic {
	return &Diagnostic{
		label:      name,
		identifier: uuid.NewString(),
		date:       time.Now(),
		inerror:    false,
		events:     []*DiagnosticEvent{},
	}
}

// NewDE is a helper function to add an invent to the diagnostic
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
// Should be called when the server asks to
func (p *Pipeline) ExecutePipeline() {
	for _, comp := range p.events {
		comp.ExecuteInPipeline(p)
	}
}
