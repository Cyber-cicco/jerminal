package jerminal

import "time"

type Agent struct {
	Identifier string
}

type Server struct {
}

type Pipeline struct {
	Agent
}

type PipelineComponent interface {
	ExecuteInPipeline()
	GetExecutionOrder() uint32
}

func SetAgent(agent Agent)

func (p *Pipeline) schedule() {

}

type Stages struct {
	Stages         []Stage
	ExecutionOrder uint32
}

func (s *Stage) Execute() error {
	beginning := time.Now().UnixMilli()
	for _, ex := range s.Executables {
        err := ex()
	}
	end := time.Now().UnixMilli()
	s.ElapsedTime = end - beginning
	return nil
}

func (s *Stages) ExecuteInPipeline() {
	for _, stage := range s.Stages {
        stage.Execute()
	}
}

func (s *Stages) GetExecutionOrder() uint32 {
	return s.ExecutionOrder
}

type Stage struct {
    Name string
	Executables []func() error
	ElapsedTime int64
}

type Executable struct {
    ShouldStopIfError bool
    Func func() error
    RecoveryFunc *Executable
}

func SetStages(stages ...Stage) *Stages {
	return &Stages{
		stages,
		0,
	}
}

func SetStage(executables ...func()) *Stage {
	return &Stage{
		Executables: executables,
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
