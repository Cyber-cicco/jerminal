package pipeline

import (
	"fmt"
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
	Diagnostic        *Diagnostic // Infos about the process
}

// executor represents a task within a stage. It includes a main executable
// and an optional recovery function to handle errors.
type executor struct {
	ex           executable // Main task to execute.
	recoveryFunc executable // Recovery task to execute in case of failure.
	deferedFunc  executable // Task to execute at the end of the stage
}

// Exec defines a function type that performs a task within a pipeline.
type Exec func(p *Pipeline) error


// Stages initializes a new set of stages with the provided configuration.
func Stages(name string, _stages ...*stage) *stages {
	return &stages{
		stages:            _stages,
		executionOrder:    0,
		shouldStopIfError: true,
		parallel:          false,
	}
}

// Execute executes a task and handles retries on failure.
func (e Exec) Execute(p *Pipeline) error {
	return e(p)
}

// Stage initializes a new stage with the provided executables.
func Stage(name string, executables ...executable) *stage {
	executors := make([]*executor, len(executables))
	for i, ex := range executables {
		switch ex.(type) {
		case *executor:
			{
				n, _ := ex.(*executor)
				executors[i] = n
			}
		case Exec:
			{
				executors[i] = &executor{
					ex:           ex,
					recoveryFunc: nil,
				}
			}
		}
	}
	return &stage{
		executors:         executors,
		shouldStopIfError: true,
		tries:             1,
		delay:             1,
	}
}

// Execute runs the tasks in a stage sequentially and records the elapsed time.
//
// TODO : implement the retries
func (s *stage) Execute(p *Pipeline) error {

	var lastErr error
	defer func() {
		for i, ex := range s.executors {
			s.Diagnostic.NewDE(DEBUG, "Executing clean up of stage")
			if ex.deferedFunc != nil {
				err := ex.deferedFunc.Execute(p)
				if err != nil {
					s.Diagnostic.NewDE(ERROR, fmt.Sprintf("Stage %s got error %v in execution n°%d", s.name, err, i))
					lastErr = err
					return
				}
			}
		}
	}()

	s.Diagnostic = NewDiag(fmt.Sprintf("%s#%d %s", p.id, s.executionOrder, s.name))

	beginning := time.Now().UnixMilli()

	s.Diagnostic.NewDE(INFO, fmt.Sprintf("Stage %s started", s.name))

	for i, ex := range s.executors {
		s.Diagnostic.NewDE(DEBUG, fmt.Sprintf("executing task n°%d of stage", i))
		err := ex.Execute(p)
		if err != nil {
			s.Diagnostic.NewDE(ERROR, fmt.Sprintf("Stage %s got error %v in execution n°%d", s.name, err, i))
			return err
		}
	}

	end := time.Now().UnixMilli()
	s.elapsedTime = end - beginning

	s.Diagnostic.NewDE(INFO, fmt.Sprintf("process %s finished in %d ms", s.name, s.elapsedTime))
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
	s.tries = retries
	return s
}

// Execute runs the executor's main task. If it fails, the recovery function is invoked.
func (e *executor) Execute(p *Pipeline) error {
	err := e.ex.Execute(p)
	if err != nil && e.recoveryFunc != nil {
		return e.recoveryFunc.Execute(p)
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

// ExecTryCatch wraps an executable with a defered function to execute at the end of the stage.
func ExecDefer(ex Exec, defered executable) executable {
	return &executor{
		ex:           ex,
		recoveryFunc: nil,
		deferedFunc:  defered,
	}
}
