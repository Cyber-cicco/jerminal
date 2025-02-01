package pipeline

import (
	"errors"
	"path/filepath"

	"github.com/Cyber-cicco/jerminal/utils"
)

// onceRunner is a stage that should execute only once for a pipeline that
// can be executed multiple times
//
// The first time it runs, it should execute the commands and caches the current
// state of the directory
//
// The subsequent runs just copies the content of the directory in the agent directory
type onceRunner struct {
	executables    []executable // List of executables to run.
	executionOrder uint32       // Order in which the executables should be executed.
	Diagnostic     *Diagnostic  // Infos about the process
}

func (o *onceRunner) GetName() string {
	return "once runnner"
}

// GetShouldStopIfError should always return true for a onceRunner
func (o *onceRunner) GetShouldStopIfError() bool {
	return true
}

// RunOnce initializes a OnceRunner with the specified executables.
func RunOnce(executables ...executable) *onceRunner {
	return &onceRunner{
		executables: executables,
	}
}

// ExecuteInPipeline runs all executables in a OnceRunner.
func (o *onceRunner) ExecuteInPipeline(p *Pipeline) error {

    pipelinePath := filepath.Join(p.state.PipelineDir, p.id.String())
    empty, err := utils.IsDirEmpty(p.directory)

    if err != nil {
        return err
    }

    if !empty {
        return errors.New("Agent directory should be empty when executing a task that runs once per pipeline")
    }

    if p.TimeRan > 0 {
        p.TimeRan++
        return nil
    }

    for _, ex := range o.executables {
        err := ex.Execute(p)
        if err != nil {
            return err
        }
    }
    
    err = utils.CopyDir(p.directory, pipelinePath)
    
    if err != nil {
        return err
    }

    p.TimeRan++
	return nil
}
