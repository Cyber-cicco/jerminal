package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// stage represents a single step in a pipeline.
// A stage contains executors that define tasks to be executed.
type stage struct {
	name              string      // Name of the stage.
	executors         []*executor // List of executors to run in this stage.
	shouldStopIfError bool        // Determines whether execution stops on error.
	elapsedTime       int64       // Time taken to execute the stage (in milliseconds).
	tries             uint16      // Number of times you have to try to execute the stage before accepting failure
	delay             uint16      // Delay between the tries
	executionOrder    uint32      // Execution order in the stages
}

// executor represents a task within a stage. It includes a main executable
// and an optional recovery function to handle errors.
type executor struct {
	ex           executable // Main task to execute.
	recoveryFunc executable // Recovery task to execute in case of failure.
	deferedFunc  executable // Task to execute at the end of the stage
}

// Exec defines a function type that performs a task within a pipeline.
type Exec func(p *Pipeline, ctx context.Context) error

// Execute executes a task and handles retries on failure.
func (e Exec) Execute(p *Pipeline, ctx context.Context) error {
	return e(p, ctx)
}

// Stage initializes a new stage with the provided executables.
func Stage(name string, executables ...executable) *stage {
	executors := make([]*executor, len(executables))
	for i, ex := range executables {
		switch ex.(type) {

		case *executor:
			n, _ := ex.(*executor)
			executors[i] = n

		case Exec:
			executors[i] = &executor{
				ex:           ex,
				recoveryFunc: nil,
			}
		}
	}
	return &stage{
		name:              name,
		executors:         executors,
		shouldStopIfError: true,
		tries:             1,
		delay:             1,
	}
}

// ExecuteStage runs the executables in a stage sequentially and records the elapsed time.
// If there is retries before failure
func (s *stage) ExecuteStage(p *Pipeline, ctx context.Context) error {
	diag := NewDiag(fmt.Sprintf("%s | stage %s", p.Name, s.name))
	p.Diagnostic.AddChild(diag)
	var err error
	var i uint16 = 0
	for true {
		err = s.executeOnce(p, diag, ctx)
		if err != nil && i+1 < s.tries {
			diag.NewDE(WARN, fmt.Sprintf("Task failed for the %d time, retrying in %d seconds", i+1, s.delay))
			time.Sleep(time.Duration(s.delay) * time.Second)
			i++
			continue
		}
		break
	}
	return err
}

// Runs the executables without caring about the number of tries
func (s *stage) executeOnce(p *Pipeline, diag *Diagnostic, ctx context.Context) error {
	var lastErr error
	defer func() {
		for i, ex := range s.executors {
			diag.NewDE(DEBUG, "Executing clean up of stage")
			if ex.deferedFunc != nil {
				err := ex.deferedFunc.Execute(p, ctx)
				if err != nil {
					diag.NewDE(ERROR, fmt.Sprintf("Stage %s got error %v in execution n°%d", s.name, err, i))
					lastErr = err
					return
				}
			}
		}
	}()

	beginning := time.Now().UnixMilli()

	diag.NewDE(INFO, fmt.Sprintf("Stage %s started", s.name))

	for i, ex := range s.executors {
		if ex.ex != nil {
			diag.NewDE(DEBUG, fmt.Sprintf("executing task n°%d of stage", i))
			err := ex.Execute(p, ctx)
			if err != nil {
				diag.NewDE(ERROR, fmt.Sprintf("Stage %s got error %v in execution n°%d", s.name, err, i))
				return err
			}
		}
	}

	end := time.Now().UnixMilli()
	s.elapsedTime = end - beginning

	diag.NewDE(INFO, fmt.Sprintf("process %s finished in %d ms", s.name, s.elapsedTime))
	return lastErr
}

func (s *stage) GetExecutionOrder() uint32 {
	return s.executionOrder
}

func (s *stage) SetExecutionOrder(order uint32) {
	s.executionOrder = order
}

// DontStopIfErr configures a stage to continue execution even if an error occurs.
func (s *stage) DontStopIfErr() *stage {
	s.shouldStopIfError = false
	return s
}

// Retry tells the current stage to retry x times with y seconds delay
// between each try
func (s *stage) Retry(retries, delaySeconds uint16) *stage {
	s.delay = delaySeconds
	s.tries += retries
	return s
}

// Execute runs the executor's main task. If it fails, the recovery function is invoked.
func (e *executor) Execute(p *Pipeline, ctx context.Context) error {
	err := e.ex.Execute(p, ctx)
	if err != nil && e.recoveryFunc != nil {
		return e.recoveryFunc.Execute(p, ctx)
	}
	return err
}

// ExecTryCatch wraps an executable with a recovery function to handle errors.
func ExecTryCatch(ex Exec, recovery executable) executable {
	return &executor{
		ex:           ex,
		recoveryFunc: recovery,
		deferedFunc:  nil,
	}
}

// ExecDefer wraps an executable with a defered function to execute at the end of the stage.
func ExecDefer(ex Exec, defered executable) executable {
	return &executor{
		ex:           ex,
		recoveryFunc: nil,
		deferedFunc:  defered,
	}
}

// Defer wraps an executable with a defered function to execute at the end of the stage.
func Defer(defered Exec) executable {
	return &executor{
		ex:           nil,
		recoveryFunc: nil,
		deferedFunc:  defered,
	}
}

// Cache copies a directory in the cache
func Cache(dirname string) executable {
	return Exec(func(p *Pipeline, ctx context.Context) error {
		targetPath := filepath.Join(p.directory, dirname)
		cachePath := filepath.Join(p.globalState.PipelineDir, p.Id.String(), dirname)
		p.Diagnostic.NewDE(INFO, fmt.Sprintf("Caching directory %s", targetPath))
		_, err := os.Stat(targetPath)

		if err != nil {
			return err
		}

		_, err = os.Stat(cachePath)

		if err != nil {
			err = os.MkdirAll(cachePath, os.ModePerm)
			if err != nil {
				return err
			}
		}

		sh := SH("rsync", targetPath, cachePath)
		err = sh.Execute(p, ctx)

		return err
	})
}
