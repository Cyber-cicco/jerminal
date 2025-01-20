package jerminal

import (
	"errors"
	"testing"
	"time"

	"github.com/Cyber-cicco/jerminal/state"
	"github.com/google/uuid"
)

func _test_getPipeline() *Pipeline {
	return &Pipeline{
		Agent:         &state.Agent{},
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

func TestStagesExecute1(t *testing.T) {
	p := _test_getPipeline()
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
    	executionOrder:    0,
    	name:              "stages",
    	stages:            []*stage{
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
	p := _test_getPipeline()
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
    	executionOrder:    0,
    	name:              "stages",
    	stages:            []*stage{
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

func TestStagesExecute3 (t *testing.T) {
    return
	p := _test_getPipeline()
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
    	executionOrder:    0,
    	name:              "stages1",
    	stages:            []*stage{
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
