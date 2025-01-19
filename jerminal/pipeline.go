package jerminal

// Pipeline represents the main execution context for stages and executors.
// It uses an Agent to manage execution and a directory for workspace.
type Pipeline struct {
	agent
	MainDirectory string              // Base directory of the pipeline
	Directory     string              // Working directory for the pipeline.
	events    []pipelineEvents // components to be executed
}

// Launches the events of the pipeline
//
// Should be called when the server asks to
func (p *Pipeline) ExecutePipeline() {
    for _, comp  := range p.events {
        comp.ExecuteInPipeline(p)
    }
}
