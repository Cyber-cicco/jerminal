package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSH(t *testing.T) {

    p := _test_getPipeline("TestSH")
    agentPath := filepath.Join(p.globalState.AgentDir, p.Agent.Identifier)
    p.mainDirectory = agentPath
    p.directory = agentPath
    os.MkdirAll(agentPath, os.ModePerm)
    defer os.RemoveAll(agentPath)

    fun1 := SH("mkdir", "test")
    fun2 := SH("rmdir", "test")
    fun1.Execute(p, context.Background())
    _, err := os.Stat(filepath.Join(p.directory, "test"))
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    fun2.Execute(p, context.Background())

    _, err = os.Stat(filepath.Join(p.directory, "test"))
    if err == nil {
        t.Fatalf("Expected an error but got nothing instead")
    }

}

func TestCD(t *testing.T) {
    p := _test_getPipeline("TestCD")
    agentPath := filepath.Join(p.globalState.AgentDir, p.Agent.Identifier)
    p.mainDirectory = agentPath
    p.directory = agentPath
    os.MkdirAll(agentPath, os.ModePerm)
    defer os.RemoveAll(agentPath)

    sh1 := SH("mkdir", "test")
    sh2 := SH("touch", "me")

    cd := CD("test")

    sh1.Execute(p, context.Background())
    err := cd.Execute(p, context.Background())

    if err != nil {
        t.Fatalf("Got an error when changing directory %v", err)
    }

    sh2.Execute(p, context.Background())

    _, err = os.Stat(filepath.Join(p.mainDirectory, "test", "me"))

    if err != nil {
        t.Fatalf("File was not created at the right place")
    }

    _, err = os.Stat(filepath.Join(p.directory, "me"))

    if err != nil {
        t.Fatalf("File was not created at the right place")
    }

    cd.deferedFunc.Execute(p, context.Background())

    sh2.Execute(p, context.Background())

    _, err = os.Stat(filepath.Join(p.mainDirectory, "me"))

    if err != nil {
        t.Fatalf("File was not created at the right place")
    }

    _, err = os.Stat(filepath.Join(p.directory, "me"))

    if err != nil {
        t.Fatalf("File was not created at the right place")
    }

}
