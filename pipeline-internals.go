package jerminal

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/Cyber-cicco/jerminal/state"
	"github.com/google/uuid"
)

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
	Diagnostic        *Diagnostic // Infos about the process
}

// executor represents a task within a stage. It includes a main executable
// and an optional recovery function to handle errors.
type executor struct {
	ex           executable // Main task to execute.
	recoveryFunc executable // Recovery task to execute in case of failure.
	deferedFunc  executable // Task to execute at the end of the stage
}

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

// Exec defines a function type that performs a task within a pipeline.
type Exec func(p *Pipeline) error

// Provides the agent for the pipeline
type AgentProvider func(p *Pipeline) *state.Agent

// pipelineEvents represents a generic event of the pipeline.
// Each event must be able to execute within a pipeline and provide metadata.
//
// Implemented by : stages, onceRunner
type pipelineEvents interface {
	ExecuteInPipeline(p *Pipeline) error // Executes the component within the pipeline.
	GetExecutionOrder() uint32           // Returns the execution order of the event.
	SetExecutionOrder(uint32)            // Sets the execution order of the event
	GetShouldStopIfError() bool          // Indicates if the pipeline should stop on error.
	GetName() string
}

// executable represents an entity that can be executed within a pipeline.
//
// Implemented by stage, Exec, executor
type executable interface {
	Execute(p *Pipeline) error // Executes the entity.
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

// ExecuteInPipeline executes all the stages within the pipeline.
func (s *stages) ExecuteInPipeline(p *Pipeline) error {

	diag := NewDiag(fmt.Sprintf("%s#%s", p.name, s.name))
	s.diagnostic = diag
	beginning := time.Now().UnixMilli()
	diag.NewDE(INFO, fmt.Sprintf("stage %s started", s.name))

	defer func() {
		end := time.Now().UnixMilli()
		elapsedTime := end - beginning

		diag.NewDE(INFO, fmt.Sprintf("stage %s ended successfully. Took %d ms", s.name, elapsedTime))
	}()

	// Parallel execution of pipelines
	if s.parallel {
		diag.NewDE(DEBUG, "starting parallel tasks")
		var wg sync.WaitGroup
		errchan := make(chan error, len(s.stages))
		for _, s := range s.stages {
			wg.Add(1)
			go func(p *Pipeline, s *stage) {
				defer wg.Done()
				err := s.Execute(p)
				if err != nil {
					if s.shouldStopIfError {
						errchan <- err
						return
					}
					diag.Lock()
					defer diag.Unlock()
					diag.NewDE(WARN, fmt.Sprintf("got non blocking error in stage %s : %v", s.name, err))
				}
			}(p, s)
		}
		wg.Wait()
		close(errchan)
		for err := range errchan {

			// returns first error encountered in the channel
			// maybe change that
			if err != nil {
				diag.NewDE(DEBUG, fmt.Sprintf("encountered error in one of the tasks. %v", err))
				return err
			}
		}
        return nil

	}

	for _, stage := range s.stages {
		err := stage.Execute(p)
		if err != nil {
			if stage.shouldStopIfError {
				return err
			}
			diag.NewDE(WARN, fmt.Sprintf("got non blocking error in stage %s : %v", s.name, err))
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
func Agent(id string) AgentProvider {
	return func(p *Pipeline) *state.Agent {
		return p.State.GetAgent(id)
	}
}

// SetPipeline initializes a new pipeline with the specified agent and components.
//
// It gets the current state of the app and gives back the Pipeline
func SetPipeline(name string, agent AgentProvider, events ...pipelineEvents) (*Pipeline, error) {
	s, err := state.GetState()
	if err != nil {
		return nil, err
	}
	return setPipelineWithState(name, agent, s, events...), nil
}

// setPipelineWithState gets a new pipeline with a state
//
// Only in testing should it be used by something else than SetPipeline
func setPipelineWithState(name string, agent AgentProvider, state *state.ApplicationState, events ...pipelineEvents) *Pipeline {
	p := Pipeline{
		name:          name,
		id:            uuid.New(),
		mainDirectory: "",
		directory:     "",
		events:        events,
		Diagnostic:    &Diagnostic{},
		timeRan:       0,
		State:         state,
	}
	p.Agent = agent(&p)
	return &p
}

// ExecuteInPipeline runs all executables in a OnceRunner.
func (o *onceRunner) ExecuteInPipeline(p *Pipeline) error {

    pipelinePath := filepath.Join(p.State.PipelineDir, p.id.String())
    fmt.Printf("p.directory: %v\n", p.directory)
    empty, err := isDirEmpty(p.directory)

    if err != nil {
        return err
    }

    if !empty {
        return errors.New("Agent directory should be empty when executing a task that runs once per pipeline")
    }

    if p.timeRan > 0 {
        err := copyDir(pipelinePath, p.directory)
        if err != nil {
            return err
        }
        p.timeRan++
        return nil
    }

    for _, ex := range o.executables {
        err := ex.Execute(p)
        if err != nil {
            return err
        }
    }
    
    err = copyDir(p.directory, pipelinePath)
    
    if err != nil {
        return err
    }

    p.timeRan++
	return nil
}

func (o *onceRunner) GetName() string {
	return "once runnner"
}

// GetShouldStopIfError should always return true for a onceRunner
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
