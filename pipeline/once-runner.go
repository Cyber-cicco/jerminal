package pipeline

import (
	"context"
	"errors"

	"github.com/Cyber-cicco/jerminal/utils"
)

// onceRunner is a stage that should execute only once for a pipeline that
// can be executed multiple times
//
// The first time it runs, it should execute the commands and caches the current
// config of the directory
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
func (o *onceRunner) ExecuteInPipeline(p *Pipeline, ctx context.Context) error {
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

	p.Diagnostic.LogEvent(INFO, "Executing pipeline setup for subsequent runs")

	for _, ex := range o.executables {
		select {
		case <-ctx.Done():
            p.Diagnostic.LogEvent(WARN, "Job got canceled before finishing")
        default:
			err := ex.Execute(p, ctx)
			if err != nil {
				return err
			}

		}
	}

	err = utils.CopyDir(p.directory, p.pipelineDir)

	if err != nil {
		return err
	}

	p.TimeRan++
	return nil
}
