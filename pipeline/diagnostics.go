package pipeline

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const DATE_TIME_LAYOUT = "2006-01-02 15:04:05"
const FILE_DATE_TIME_LAYOUT = "2006-01-02_15-04-05"

var IMPORTANCE_STR = [5]string{"DEBUG", "INFO", "WARN", "ERROR", "CRITICAL"}

// Informations about an element of the pipeline
type Diagnostic struct {
	Start        JSONTime      `json:"date" time_format:"2006-01-02 15:04:05"` // Time the diagnostic was written
	Events       []pipelineLog `json:"logs"`                                   // Infos about what happened in the process
	Label        string        `json:"label"`                                  // Name of the diagnostic
	identifier   uuid.UUID     `json:"-"`                                      // Unique identifier of the diagnostic
	parent       *Diagnostic   `json:"-"`                                      // Parent of the Diagnostic. Nil if does not exist
	sync.RWMutex `json:"-"`    // Can be used in goroutines so need to lock it
	Inerror      bool          `json:"in-error"` // Tells if the attached process should be considered in error
}

// Infos about an event
type DiagnosticEvent struct {
	Description string      `json:"description"` // Description of the event
	Time        string      `json:"time"`        // Time of the event happening
	Name        string      `json:"name"`        // Name to display in the log
	Importance  EImportance `json:"importance"`  // Importance of the event
}

// LogEvent is a helper function to add an event to the diagnostic
func (d *Diagnostic) LogEvent(importance EImportance, description string) {
	d.Lock()
	defer d.Unlock()
	newEvt := &DiagnosticEvent{
		Importance:  importance,
		Description: description,
		Time:        time.Now().Format(DATE_TIME_LAYOUT),
		Name:        d.Label,
	}
	newEvt.Log()
	d.Events = append(d.Events, newEvt)
}

// Creates a new diag with a filter based on importance
//
// TODO : inefficient, remove cloning, and implement it on a Marshalling level
func (d *Diagnostic) FilterBasedOnImportance(imp EImportance) *Diagnostic {
	newDiag := &Diagnostic{
		Label:      d.Label,
		identifier: d.identifier,
		Start:      d.Start,
		Inerror:    d.Inerror,
		parent:     d.parent,
		Events:     []pipelineLog{},
	}
	for _, ev := range d.Events {
		switch e := ev.(type) {

		case *DiagnosticEvent:
			if e.Importance >= imp {
				newDiag.Events = append(newDiag.Events, e)
			}

		case *Diagnostic:
			newDiag.Events = append(newDiag.Events, e.FilterBasedOnImportance(imp))

		}
	}
	return newDiag
}

// Adds a child diag to the current diagnostic
func (d *Diagnostic) AddChild(diag *Diagnostic) {
	d.Lock()
	defer d.Unlock()
	d.Events = append(d.Events, diag)
	diag.parent = d
}

// Prints the events recursively
func (d *Diagnostic) Log() {
	for _, ev := range d.Events {
		ev.Log()
	}
}

// Logs the event in standard output. Might want to have other options as well
func (d *DiagnosticEvent) Log() {
	fmt.Printf("[%s] - %s at %s: %s\n", IMPORTANCE_STR[d.Importance], d.Name, d.Time, d.Description)
}

// NewDiag gets a Diagnostic with defaults
func NewDiag(name string) *Diagnostic {
	return &Diagnostic{
		Label:      name,
		identifier: uuid.New(),
		Start:      JSONTime(time.Now()),
		Inerror:    false,
		Events:     []pipelineLog{},
	}
}
