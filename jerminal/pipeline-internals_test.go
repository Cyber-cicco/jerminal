package jerminal

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func _test_getPipeline() *Pipeline {
	return &Pipeline{
		agent:         agent{},
		name:          "test",
		mainDirectory: "./test",
		directory:     "./test",
		id:            uuid.New(),
		timeRan:       0,
		events:        []pipelineEvents{},
		inerror:       false,
		Diagnostic:    &Diagnostic{},
	}
}

func TestStageExecute1(t *testing.T) {
	p := _test_getPipeline()
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
	p := _test_getPipeline()
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

	if err == nil {
		t.Fatalf("Expected error nothing instead")
	}

	if actual != expected {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestExecTryCatch(t *testing.T) {
	p := _test_getPipeline()
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
