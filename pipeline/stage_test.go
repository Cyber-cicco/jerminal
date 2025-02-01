package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStageExecute1(t *testing.T) {
	p := _test_getPipeline("TestStageExecute1")
	actual := 0
	stage := stage{
		name: "test1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("first exec func")
					actual++
					return nil
				}),
				recoveryFunc: nil,
				deferedFunc: Exec(func(p *Pipeline) error {
					t.Log("defered exec func")
					actual *= actual
					return nil
				}),
			},
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second exec func")
					actual++
					return nil
				}),
				recoveryFunc: nil,
				deferedFunc:  nil,
			},
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("third exec func")
					actual++
					return errors.New("test")
				}),
				recoveryFunc: Exec(func(p *Pipeline) error {
					t.Log("forth exec func")
					actual++
					return nil
				}),
				deferedFunc: nil,
			},
		},
		shouldStopIfError: true,
		elapsedTime:       0,
		tries:             0,
		delay:             0,
		executionOrder:    0,
	}

	err := stage.ExecuteStage(p)

	if err != nil {
		t.Fatalf("Did not expect error, got %v", err)
	}

	expected := 16

	if actual != expected {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestStageExecute2(t *testing.T) {
	p := _test_getPipeline("TestStageExecute2")
	actual := 0
	expected := 9

	stage := stage{
		name: "test1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("first exec func")
					actual++
					return nil
				}),
				recoveryFunc: nil,
				deferedFunc: Exec(func(p *Pipeline) error {
					t.Log("defered exec func")
					actual *= actual
					return nil
				}),
			},
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second exec func")
					actual++
					return nil
				}),
				recoveryFunc: nil,
				deferedFunc:  nil,
			},
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("third exec func")
					actual++
					return errors.New("test")
				}),
				recoveryFunc: nil,
				deferedFunc:  nil,
			},
		},
		shouldStopIfError: true,
		elapsedTime:       0,
		tries:             0,
		delay:             0,
		executionOrder:    0,
	}

	err := stage.ExecuteStage(p)

	if err == nil {
		t.Fatalf("Expected error nothing instead")
	}

	if actual != expected {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestExecTryCatch(t *testing.T) {
	p := _test_getPipeline("TestExecTryCatch")
	err := errors.New("test")
	actual := 0
	expected := 4
	exec := ExecTryCatch(
		func(p *Pipeline) error {
			actual++
			return err
		},
		ExecTryCatch(
			func(p *Pipeline) error {
				actual++
				return err
			},
			ExecTryCatch(
				func(p *Pipeline) error {
					actual++
					return err
				},
				Exec(func(p *Pipeline) error {
					actual++
					return err
				}),
			),
		),
	)
	err = exec.Execute(p)
	if err == nil {
		t.Fatalf("Expected an error, got nothing instead")
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}

}


func TestCache(t *testing.T) {
    p := _test_getPipeline("TestCache")
    cache := Cache("test")
    agentPath := filepath.Join(p.state.AgentDir, p.agent.Identifier)
    pipeLinePath := filepath.Join(p.state.PipelineDir, p.id.String())
    p.mainDirectory = agentPath
    p.directory = agentPath
    os.MkdirAll(filepath.Join(p.directory, "test"), os.ModePerm)
    cache.Execute(p)
    _, err := os.Stat(filepath.Join(pipeLinePath, "test"))

    if err != nil {
        t.Fatalf("Expected no error, got %s", err)
    }
}
