package pipeline

type DEImp uint8

const (
	DEBUG = DEImp(iota)
	INFO
	WARN
	ERROR
	CRITICAL
)

// pipelineEvents represents a generic event of the pipeline.
// Each event must be able to execute within a pipeline and provide metadata.
//
// Implemented by : stages, onceRunner
type pipelineEvents interface {
	ExecuteInPipeline(p *Pipeline) error // Executes the component within the pipeline.
	GetShouldStopIfError() bool          // Indicates if the pipeline should stop on error.
	GetName() string
}

// Can be Diagnostic or DiagEvent
type pipelineLog interface {
    Log()
}

// executable represents an entity that can be executed within a pipeline.
//
// Implemented by Exec, executor
type executable interface {
	Execute(p *Pipeline) error // Executes the entity.
}
