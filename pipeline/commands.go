package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CD restricts navigation to prevent access to parent directories or absolute paths.
//
// It gives back two functions : one to set the context of the stage to the specified
// diretory, and one to execute later that gets the puts the current directory back to the original
//
// It should be used with the ExecDefer function
func CD(dir string) *executor {
	cd := func(p *Pipeline) error {
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

	defered := func(p *Pipeline) error {
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
func SH(name string, args ...string) executable {
	return Exec(func(p *Pipeline) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = p.directory
        p.Diagnostic.NewDE(DEBUG, fmt.Sprintf("Executing command %s", name))
		out, err := cmd.CombinedOutput()
        p.Diagnostic.NewDE(DEBUG, fmt.Sprintf("Got ouput : %s", string(out)))
		return err
	})
}
