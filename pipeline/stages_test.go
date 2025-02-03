package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStagesExecute1(t *testing.T) {
	p := _test_getPipeline("TestStagesExecute1")
	actual := 0
	expected := 2
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
					t.Log("second stage func")
					actual++
					return nil
				}),
			},
		},
	}
	stages := stages{
		name:           "stages",
		stages: []*stage{
			&stage1, &stage2,
		},
		shouldStopIfError: true,
		parallel:          false,
	}
	err := stages.ExecuteInPipeline(p, context.Background())

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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
					t.Log("second stage func")
					actual++
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stages := stages{
		name:           "stages",
		stages: []*stage{
			&stage1, &stage2, &stage3, &stage4,
		},
		shouldStopIfError: false,
		parallel:          false,
	}
	err := stages.ExecuteInPipeline(p, context.Background())

	if err == nil {
		t.Fatalf("Expected an error, got nothing")
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestStagesExecute3(t *testing.T) {
	p := _test_getPipeline("TestStagesExecute3")
	stage1 := stage{
		name: "stage1",
		executors: []*executor{
			{
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(1 * time.Second)
					return nil
				}),
			},
		},
		shouldStopIfError: true,
	}
	stages1 := stages{
		name:           "stages1",
		stages: []*stage{
			&stage1, &stage2, &stage3,
		},
		shouldStopIfError: false,
		parallel:          false,
	}
	begin := time.Now().Unix()
	err := stages1.ExecuteInPipeline(p, context.Background())
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
	err = stages1.ExecuteInPipeline(p, context.Background())
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
				ex: Exec(func(p *Pipeline, ctx context.Context) error {
					return errors.New("test")
				}),
			},
		},
		shouldStopIfError: false,
        tries: 2,
        delay: 1,
	}
    begin := time.Now().Unix()
    err := stage1.ExecuteStage(p, context.Background())
    end := time.Now().Unix()

    if err == nil {
        t.Fatalf("Expected error, but it worked instead")
    }
    delay := end - begin
    if delay != 1 {
        t.Fatalf("Expected 1, got %d", delay)
    }
}

