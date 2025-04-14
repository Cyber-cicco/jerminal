package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestOnceRunner(t *testing.T) {
	p := _test_getPipeline("TestOnceRunner")
	o := &onceRunner{
		executables: []executable{
			Exec(func(p *Pipeline, ctx context.Context) error {
				err := os.Mkdir(filepath.Join(p.directory, "test"), os.ModePerm)
				return err
			}),
		},
		executionOrder: 0,
		Diagnostic:     &Diagnostic{},
	}

	dirPathPipe := filepath.Join(p.globalState.PipelineDir, p.Id.String())
	dirPathAgent := filepath.Join(filepath.Join(p.globalState.AgentDir, p.Agent.Identifier))
    fmt.Printf("dirPathAgent: %v\n", dirPathAgent)
	os.MkdirAll(dirPathPipe, os.ModePerm)
	os.MkdirAll(dirPathAgent, os.ModePerm)
	defer func() {
		os.RemoveAll(dirPathPipe)
		os.RemoveAll(dirPathAgent)
	}()

	p.mainDirectory = dirPathAgent
    p.pipelineDir = dirPathPipe
	p.directory = dirPathAgent

	err := o.ExecuteInPipeline(p, context.Background())

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

}
