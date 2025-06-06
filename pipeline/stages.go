package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Stages represents a collection of pipeline stages.
// Each stage has an execution order and can be configured to stop if an error occurs.
type stages struct {
	name              string   // The identifier of the stages
	stages            []*stage // List of stages in the pipeline.
	shouldStopIfError bool     // Determines whether execution should stop on error.
	parallel          bool     // Determines wether execution of stages should be put in goroutines
}

// Stages initializes a new set of stages to execute in sequence by default.
// Can be made parallel with the Parallel() function
func Stages(name string, _stages ...*stage) *stages {
	return &stages{
		name:              name,
		stages:            _stages,
		shouldStopIfError: true,
		parallel:          false,
	}
}

// ExecuteInPipeline executes all the stages within the pipeline.
func (s *stages) ExecuteInPipeline(p *Pipeline, ctx context.Context) error {

	diag := NewDiag(fmt.Sprintf("%s | stages %s", p.Name, s.name))
	p.Diagnostic.AddChild(diag)
	p.Diagnostic = diag
	beginning := time.Now().UnixMilli()
	diag.LogEvent(INFO, fmt.Sprintf("stages %s started", s.name))

	defer func() {
		end := time.Now().UnixMilli()
		elapsedTime := end - beginning
		diag.LogEvent(INFO, fmt.Sprintf("stages %s ended successfully. Took %d ms", s.name, elapsedTime))
		p.ResetDiag()
	}()

	// Parallel execution of pipelines
	// Parallel execution seem to pose a problem with diags in stages
	if s.parallel {
		diag.LogEvent(DEBUG, "starting parallel tasks")
		var wg sync.WaitGroup
		errchan := make(chan error, len(s.stages))
		for _, s := range s.stages {
			wg.Add(1)
			go func(p *Pipeline, s *stage) {
				defer wg.Done()
				err := s.ExecuteStage(p, ctx)
				if err != nil {
					if s.shouldStopIfError {
						errchan <- err
						return
					}
					diag.LogEvent(WARN, fmt.Sprintf("got non blocking error in stage %s : %v", s.name, err))
				}
			}(p, s)
		}
		wg.Wait()
		close(errchan)
		for err := range errchan {

			// returns first error encountered in the channel
			// maybe change that
			if err != nil {
				diag.LogEvent(DEBUG, fmt.Sprintf("encountered error in one of the tasks. %v", err))
				return err
			}
		}
		return nil

	}

	for _, stage := range s.stages {
		select {
		case <-ctx.Done():
			diag.LogEvent(WARN, fmt.Sprintf("Stages got canceled before finishing"))
			return ctx.Err()
		default:
			err := stage.ExecuteStage(p, ctx)
			if err != nil {
				if stage.shouldStopIfError {
					return err
				}
				diag.LogEvent(WARN, fmt.Sprintf("got non blocking error in stage %s : %v", s.name, err))
			}
		}
	}
	return nil
}

func (s *stages) GetName() string {
	return s.name
}

// Parallel activates the parallel execution of stages
func (s *stages) Parallel() *stages {
	s.parallel = true
	return s
}

// GetShouldStopIfError returns whether the pipeline should stop if an error occurs in a stage.
func (s *stages) GetShouldStopIfError() bool {
	return s.shouldStopIfError
}
