package jerminal

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// agent represents a process that executes a pipeline in its personal directory
// and cleans it up afterward. The Identifier uniquely identifies the agent.
type agent struct {
	identifier string
}

// Stages represents a collection of pipeline stages.
// Each stage has an execution order and can be configured to stop if an error occurs.
type stages struct {
	executionOrder    uint32      // Order in which the stages should be executed.
	name              string      // The identifier of the stages
	stages            []*stage    // List of stages in the pipeline.
	shouldStopIfError bool        // Determines whether execution should stop on error.
	parallel          bool        // Determines wether execution of stages should be put in goroutines
	diagnostic        *Diagnostic // Infos about the process
}

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
	Diagnostic        *Diagnostic
}

// executor represents a task within a stage. It includes a main executable
// and an optional recovery function to handle errors.
type executor struct {
	ex           executable // Main task to execute.
	recoveryFunc executable // Recovery task to execute in case of failure.
	deferedFunc  executable // Task to execute at the end of the stage
}

// onceRunner is a utility for executing a set of executables once, in a specific order.
type onceRunner struct {
	executables    []executable // List of executables to run.
	executionOrder uint32       // Order in which the executables should be executed.
}

// Exec defines a function type that performs a task within a pipeline.
// It returns an error if the task fails.
type Exec func(p *Pipeline) error

// pipelineEvents represents a generic event of the pipeline.
// Each event must be able to execute within a pipeline and provide metadata.
type pipelineEvents interface {
	ExecuteInPipeline(p *Pipeline) error // Executes the component within the pipeline.
	GetExecutionOrder() uint32           // Returns the execution order of the event.
	SetExecutionOrder(uint32)            // Sets the execution order of the event
	GetShouldStopIfError() bool          // Indicates if the pipeline should stop on error.
}

// executable represents an entity that can be executed within a pipeline.
type executable interface {
	Execute(p *Pipeline) error // Executes the entity.
}

// SetAgent initializes the agent for a pipeline.
func SetAgent(ag agent) {
	// Placeholder for setting up an agent.
}

// schedule is a private method that handles pipeline scheduling logic.
func (p *Pipeline) schedule() {
	// Placeholder for scheduling pipeline execution.
}

// Execute runs the tasks in a stage sequentially and records the elapsed time.
//
// TODO : implement the retries
func (s *stage) Execute(p *Pipeline) error {

	s.Diagnostic = NewDiag(fmt.Sprintf("%s#%d %s", p.identifier, s.executionOrder, s.name))

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

	for i, ex := range s.executors {
		s.Diagnostic.NewDE(DEBUG, "Executing clean up of stage")
		if ex.deferedFunc != nil {
			err := ex.deferedFunc.Execute(p)
			if err != nil {
				s.Diagnostic.NewDE(ERROR, fmt.Sprintf("Stage %s got error %v in execution n°%d", s.name, err, i))
				return err
			}
		}
	}

	s.Diagnostic.NewDE(INFO, fmt.Sprintf("Executing clean up of", s.name, beginning))
	end := time.Now().UnixMilli()
	s.elapsedTime = end - beginning

	return nil
}

func (s *stage) GetExecutionOrder() uint32 {
	return s.executionOrder
}

func (s *stage) SetExecutionOrder(order uint32) {
	s.executionOrder = order
}

// ExecuteInPipeline executes all the stages within the pipeline.
func (s *stages) ExecuteInPipeline(p *Pipeline) error {
	diag := NewDiag(fmt.Sprintf("%s#%s", p.name, s.name))
	s.diagnostic = diag

	beginning := time.Now().UnixMilli()

	diag.NewDE(INFO, fmt.Sprintf("stage %s started", s.name))

	if s.parallel {
        diag.NewDE(DEBUG, fmt.Sprintf("starting parallel tasks", s.name))
		var wg sync.WaitGroup
		errchan := make(chan error, len(s.stages))
		for _, stage := range s.stages {
			wg.Add(1)
			go func(p *Pipeline) {
				defer wg.Done()
				errchan <- stage.Execute(p)
			}(p)
		}
		wg.Wait()
		close(errchan)
        for err := range errchan {
            if err != nil {
                diag.NewDE(DEBUG, fmt.Sprintf("encountered error in one of the tasks. %v", err))
                return err
            }
        }
	} else {
		for _, stage := range s.stages {
			err := stage.Execute(p)
			if err != nil {
				return err
			}
		}
	}

	end := time.Now().UnixMilli()
	elapsedTime := end - beginning

	diag.NewDE(INFO, fmt.Sprintf("stage %s ended successfully. Took %d ms", s.name, elapsedTime))

	return nil
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

// GetExecutionOrder returns the execution order of the stages.
func (s *stages) GetExecutionOrder() uint32 {
	return s.executionOrder
}

func (s *stages) SetExecutionOrder(order uint32) {
	s.executionOrder = order
}

// Execute runs the executor's main task. If it fails, the recovery function is invoked.
func (e *executor) Execute(p *Pipeline) error {
	err := e.ex.Execute(p)
	if err != nil && e.recoveryFunc != nil {
		return e.recoveryFunc.Execute(p)
	}
	return err
}

// Stages initializes a new set of stages with the provided configuration.
func Stages(name string, _stages ...*stage) *stages {
	return &stages{
		stages:            _stages,
		executionOrder:    0,
		shouldStopIfError: true,
		parallel:          false,
	}
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

// Execute executes a task and handles retries on failure.
func (e Exec) Execute(p *Pipeline) error {
	return e(p)
}

// Retry retries the execution of a task up to a maximum number of attempts with delays.
func (e Exec) Retry(p *Pipeline) error {
	maxRetries := 5
	delay := 3 * time.Second
	var err error

	for i := 1; i <= maxRetries; i++ {
		err = e.Execute(p)
		if err == nil {
			break
		}

		// If this was the last attempt, log the failure and exit.
		if i == maxRetries {
			log.Printf("Task failed after %d attempts: %v\n", i, err)
			break
		}

		log.Printf("Retrying in %v... (attempt %d/%d)\n", delay, i, maxRetries)
		time.Sleep(delay) // Wait before the next attempt.
	}
	log.Println("Done.")
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

// Agent retrieves an agent with the specified identifier.
func Agent(id string) agent {
	return agent{
		identifier: id,
	}
}

// SetPipeline initializes a new pipeline with the specified agent and components.
func SetPipeline(agent agent, components ...pipelineEvents) Pipeline {
	return Pipeline{}
}

// ExecuteInPipeline runs all executables in a OnceRunner.
func (o *onceRunner) ExecuteInPipeline(p *Pipeline) error {
	// TODO : implement function body
	return nil
}

// GetShouldStopIfError indicates if the OnceRunner should stop on error.
func (o *onceRunner) GetShouldStopIfError() bool {
	return true
}

// GetExecutionOrder returns the execution order for the OnceRunner.
func (o *onceRunner) GetExecutionOrder() uint32 {
	return o.executionOrder
}

func (o *onceRunner) SetExecutionOrder(order uint32) {
	o.executionOrder = order
}

// RunOnce initializes a OnceRunner with the specified executables.
func RunOnce(executables ...executable) *onceRunner {
	return &onceRunner{
		executables: executables,
	}
}
