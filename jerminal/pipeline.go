package jerminal

import (
	"log"
	"time"
)

type Agent struct {
	Identifier string
}


type Stages struct {
	Stages         []*Stage
	ExecutionOrder uint32
    ShouldStopIfError bool
}

type Pipeline struct {
	Agent
    Directory string
}

type Stage struct {
	Name              string
	Executors         []*Executor
	ShouldStopIfError bool
	ElapsedTime       int64
}

type Executor struct {
	ex           Executable
	RecoveryFunc Executable
}

type OnceRunner struct {
    executables []Executable
    ExecutionOrder uint32
}

type Exec func(p *Pipeline) error

type PipelineComponent interface {
	ExecuteInPipeline(p *Pipeline)
	GetExecutionOrder() uint32
    GetShouldStopIfError() bool
}

type Executable interface {
	Execute(p *Pipeline) error
}

func SetAgent(agent Agent) {

}

func (p *Pipeline) schedule() {

}

func (s *Stage) Execute(p *Pipeline) error {
	beginning := time.Now().UnixMilli()
	for _, ex := range s.Executors {
		err := ex.Execute(p)
		if err != nil {
			return err
		}
	}
	end := time.Now().UnixMilli()
	s.ElapsedTime = end - beginning
	return nil
}

func (s *Stages) ExecuteInPipeline(p *Pipeline) {
	for _, stage := range s.Stages {
		stage.Execute(p)
	}
}

func (s *Stages) GetShouldStopIfError() bool {
    return s.ShouldStopIfError
}

func (s *Stages) GetExecutionOrder() uint32 {
	return s.ExecutionOrder
}

func (e *Executor) Execute(p *Pipeline) error {
	err := e.Execute(p)
	if err != nil && e.RecoveryFunc != nil {
		return e.RecoveryFunc.Execute(p)
	}
	return err
}

func SetStages(stages ...*Stage) *Stages {
	return &Stages{
		Stages:            stages,
		ExecutionOrder:    0,
		ShouldStopIfError: true,
	}
}

func SetStage(name string, executables ...Executable) *Stage {
	executors := make([]*Executor, len(executables))
	for i, ex := range executables {
		switch ex.(type) {
		case *Executor:
			{
				n, _ := ex.(*Executor)
				executors[i] = n
			}
		case Exec:
			{
				executors[i] = &Executor{
					ex:           ex,
					RecoveryFunc: nil,
				}
			}
		}
	}
	return &Stage{
		Executors:         executors,
		ShouldStopIfError: true,
	}
}

func (s *Stage) DontStopIfErr() *Stage {
	s.ShouldStopIfError = false
	return s
}

func (e Exec) Execute(p *Pipeline) error {
	return e(p)
}

func (e Exec) Retry(p *Pipeline) {
	maxRetries := 5
	delay := 3 * time.Second

	for i := 1; i <= maxRetries; i++ {
		err := e.Execute(p)
		if err == nil {
			break
		}

		// If this was the last attempt, log the failure and exit
		if i == maxRetries {
			log.Printf("Task failed after %d attempts: %v\n", i, err)
			break
		}

		log.Printf("Retrying in %v... (attempt %d/%d)\n", delay, i, maxRetries)
		time.Sleep(delay) // Wait before the next attempt
	}
	log.Println("Done.")
}

func ExecTryCatch(ex Exec, recovery Executable) Executable {
	return &Executor{
		ex:           ex,
		RecoveryFunc: recovery,
	}
}

func GetAgent(id string) Agent {
	return Agent{
		Identifier: id,
	}
}

func SetPipeline(agent Agent, components ...PipelineComponent) Pipeline {
	return Pipeline{}
}

func (o *OnceRunner) ExecuteInPipeline(p *Pipeline) {

}

func (o *OnceRunner) GetShouldStopIfError() bool {
    return true
}

func (o *OnceRunner) GetExecutionOrder() uint32 {
    return o.ExecutionOrder
}

func RunOnce(executables ...Executable) *OnceRunner {
    return &OnceRunner{
    	executables:       executables,
    }
}
