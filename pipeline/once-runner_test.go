package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

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
