package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const DATE_TIME_LAYOUT = "2006-01-02 15:04"

var IMPORTANCE_STR = [5]string{"DEBUG", "INFO", "WARN", "ERROR", "CRITICAL"}

// Informations about an element of the pipeline
type Diagnostic struct {
	Label        string             // Name of the diagnostic
	identifier   uuid.UUID          // Unique identifier of the diagnostic
	Date         time.Time          // Time the diagnostic was written
	Inerror      bool               // Tells if the attached process should be considered in error
	Events       []*DiagnosticEvent // Infos about what happened in the process
	sync.RWMutex                    // Can be used in goroutines so need to lock it
	children     []*Diagnostic      // Diagnostics of children processes
	parent       *Diagnostic        // Parent of the Diagnostic. Nil if does not exist
}

// Infos about an event
type DiagnosticEvent struct {
	Importance  DEImp  // Importance of the event
	Description string // Description of the event
	Time        string
}

// NewDE is a helper function to add an event to the diagnostic
func (d *Diagnostic) NewDE(importance DEImp, description string) {
    d.Lock()
    defer d.Unlock()
	newEvt := &DiagnosticEvent{
		Importance:  importance,
		Description: description,
		Time:        time.Now().Format(DATE_TIME_LAYOUT),
	}
	newEvt.Log(d)
	d.Events = append(d.Events, newEvt)
}

func (d *Diagnostic) AddChild(diag *Diagnostic) {
    d.Lock()
    defer d.Unlock()
	d.children = append(d.children, diag)
    diag.parent = d
}

// Logs the event in standard output. Might want to have other options as well
func (d *DiagnosticEvent) Log(diag *Diagnostic) {
	fmt.Printf("[%s] - %s at %s: %s\n", IMPORTANCE_STR[d.Importance], diag.Label, d.Time, d.Description)
}

// NewDiag gets a Diagnostic with defaults
func NewDiag(name string) *Diagnostic {
	return &Diagnostic{
		Label:      name,
		identifier: uuid.New(),
		Date:       time.Now(),
		Inerror:    false,
		Events:     []*DiagnosticEvent{},
	}
}
