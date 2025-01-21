package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
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

// Infos about an event
type DiagnosticEvent struct {
	importance  DEImp  // Importance of the event
	description string // Description of the event
}

// NewDE is a helper function to add an event to the diagnostic
func (d *Diagnostic) NewDE(importance DEImp, description string) {
	d.events = append(d.events, &DiagnosticEvent{
		importance:  importance,
		description: fmt.Sprintf("%v : %s", time.Now(), description),
	})
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
