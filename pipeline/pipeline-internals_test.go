package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cyber-cicco/jerminal/state"
	"github.com/google/uuid"
)

func _test_getPipeline(agentId string) *Pipeline {
	return &Pipeline{
		Agent: &state.Agent{
			Identifier: agentId,
		},
		name:          "test",
		mainDirectory: "./test",
		directory:     "./test",
		id:            uuid.New(),
		timeRan:       0,
		events:        []pipelineEvents{},
		inerror:       false,
		Diagnostic:    &Diagnostic{},
		State: state.GetStateCustomConf(&state.Config{
			AgentDir:    "./test/agent",
			PipelineDir: "./test/pipeline",
            JerminalResourcePath: "../resources/jerminal.json",
		}),
	}
}

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
		Diagnostic:        &Diagnostic{},
	}

	err := stage.Execute(p)

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
		Diagnostic:        &Diagnostic{},
	}

	err := stage.Execute(p)

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

func TestStagesExecute1(t *testing.T) {
	p := _test_getPipeline("TestStagesExecute1")
	actual := 0
	expected := 2
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("first stage func")
					actual++
					return nil
				}),
			},
		},
	}
	stage2 := stage{
		name: "stage2",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second stage func")
					actual++
					return nil
				}),
			},
		},
	}
	stages := stages{
		executionOrder: 0,
		name:           "stages",
		stages: []*stage{
			&stage1, &stage2,
		},
		shouldStopIfError: true,
		parallel:          false,
		diagnostic:        &Diagnostic{},
	}
	err := stages.ExecuteInPipeline(p)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestStagesExecute2(t *testing.T) {
	p := _test_getPipeline("TestStagesExecute2")
	actual := 0
	expected := 3
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("first stage func")
					actual++
					return errors.New("test")
				}),
			},
		},
		shouldStopIfError: false,
	}
	stage2 := stage{
		name: "stage2",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second stage func")
					actual++
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stage3 := stage{
		name: "stage3",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second stage func")
					actual++
					return errors.New("test")
				}),
			},
		},
		shouldStopIfError: true,
	}
	stage4 := stage{
		name: "stage4",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					t.Log("second stage func")
					actual++
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stages := stages{
		executionOrder: 0,
		name:           "stages",
		stages: []*stage{
			&stage1, &stage2, &stage3, &stage4,
		},
		shouldStopIfError: false,
		parallel:          false,
		diagnostic:        &Diagnostic{},
	}
	err := stages.ExecuteInPipeline(p)

	if err == nil {
		t.Fatalf("Expected an error, got nothing")
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestStagesExecute3(t *testing.T) {
	return
	p := _test_getPipeline("TestStagesExecute3")
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					time.Sleep(1 * time.Second)
					return nil
				}),
			},
		},
		shouldStopIfError: false,
	}
	stage2 := stage{
		name: "stage2",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					time.Sleep(1 * time.Second)
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stage3 := stage{
		name: "stage3",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					time.Sleep(1 * time.Second)
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stages1 := stages{
		executionOrder: 0,
		name:           "stages1",
		stages: []*stage{
			&stage1, &stage2, &stage3,
		},
		shouldStopIfError: false,
		parallel:          false,
		diagnostic:        &Diagnostic{},
	}
	begin := time.Now().Unix()
	err := stages1.ExecuteInPipeline(p)
	end := time.Now().Unix()

	spent := end - begin

	if err != nil {
		t.Fatalf("Should not have been an error, got %v", err)
	}

	if spent < 3 {
		t.Fatalf("Test should have taken more than 3 seconds")
	}
	stages1.parallel = true

	begin = time.Now().Unix()
	err = stages1.ExecuteInPipeline(p)
	end = time.Now().Unix()

	spent = end - begin

	if err != nil {
		t.Fatalf("Should not have been an error, got %v", err)
	}

	if spent > 2 {
		t.Fatalf("Test should have taken more than 3 seconds")
	}

}

func TestStagesExecute4(t *testing.T) {
    p := _test_getPipeline("TestStagesExecute4")
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline) error {
					return errors.New("test")
				}),
			},
		},
		shouldStopIfError: false,
        tries: 2,
        delay: 1,
	}
    begin := time.Now().Unix()
    err := stage1.Execute(p)
    end := time.Now().Unix()

    if err == nil {
        t.Fatalf("Expected error, but it worked instead")
    }
    delay := end - begin
    if delay != 1 {
        t.Fatalf("Expected 1, got %d", delay)
    }
}

func TestOnceRunner(t *testing.T) {
	p := _test_getPipeline("TestOnceRunner")
	o := &onceRunner{
		executables: []executable{
			Exec(func(p *Pipeline) error {
				err := os.Mkdir(filepath.Join(p.directory, "test"), os.ModePerm)
				return err
			}),
		},
		executionOrder: 0,
		Diagnostic:     &Diagnostic{},
	}

	dirPathPipe := filepath.Join(p.State.PipelineDir, p.id.String())
	dirPathAgent := filepath.Join(filepath.Join(p.State.AgentDir, p.Agent.Identifier))
    os.MkdirAll(dirPathPipe, os.ModePerm)
    os.MkdirAll(dirPathAgent, os.ModePerm)

    p.mainDirectory = dirPathAgent
    p.directory = dirPathAgent

	err := o.ExecuteInPipeline(p)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

    infos, err := os.Stat(filepath.Join(dirPathAgent, "test"))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if infos.Name() != "test" {
		t.Fatalf("Expected file to be called test, got %s", infos.Name())
	}

	infos, err = os.Stat(filepath.Join(dirPathPipe, "test"))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if infos.Name() != "test" {
		t.Fatalf("Expected file to be called test, got %s", infos.Name())
	}

	os.RemoveAll(dirPathPipe)
	os.RemoveAll(dirPathAgent)
}
