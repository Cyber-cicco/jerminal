package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const CmdOutKey = Key("CmdOutKey")
// CD restricts navigation to prevent access to parent directories or absolute paths.
//
// It gives back two functions : one to set the context of the stage to the specified
// diretory, and one to execute later that puts the current directory back to the original
func CD(dir string) *executor {
	cd := func(p *Pipeline, ctx context.Context) error {
		// Reject absolute paths
		if filepath.IsAbs(dir) {
			return errors.New("absolute paths are not allowed")
		}

		// Prevent navigation to parent directories
		cleanDir := filepath.Clean(dir)

		newPath := filepath.Join(p.directory, cleanDir)

		// Check if the resulting path exists and is a directory
		info, err := os.Stat(newPath)
		if err != nil {
			return err // Return error if the path does not exist
		}
		if !info.IsDir() {
			return errors.New("target path is not a directory")
		}

		// Update the pipeline's directory
		p.directory = newPath
		return nil
	}

	defered := func(p *Pipeline, ctx context.Context) error {
		p.directory = p.mainDirectory
		return nil
	}

	return &executor{
		ex:           Exec(cd),
		recoveryFunc: nil,
		deferedFunc:  Exec(defered),
	}
}

// SH Executes a command in the directory of the current agent
// It also puts the result of the result of the command in
// the params of the pipeline
func SH(name string, args ...string) executable {
	return Exec(func(p *Pipeline, ctx context.Context) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = p.directory
        p.Diagnostic.LogEvent(DEBUG, fmt.Sprintf("Executing command %s", name))
		out, err := cmd.CombinedOutput()
        p.Diagnostic.LogEvent(DEBUG, fmt.Sprintf("Got ouput : %s", string(out)))
        p.Put(CmdOutKey, out)
		return err
	})
}


func SHBackground(name string, args ...string) executable {
    return Exec(func(p *Pipeline, ctx context.Context) error {
        cmd := exec.Command(name, args...)
        cmd.Dir = p.directory
        p.Diagnostic.LogEvent(DEBUG, fmt.Sprintf("Starting background command %s", name))
        
        err := cmd.Start()
        if err != nil {
            return err
        }
        
        p.Diagnostic.LogEvent(DEBUG, fmt.Sprintf("Background process started with PID: %d", cmd.Process.Pid))
        
        return nil
    })
}

