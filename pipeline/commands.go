package pipeline

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"

)

// CD restricts navigation to prevent access to parent directories or absolute paths.
//
// It gives back two functions : one to set the context of the stage to the specified
//diretory, and one to execute later that gets the puts the current directory back to the original
//
// It should be used with the ExecDefer function
func CD(dir string) (func(p *Pipeline) error, executable){
	return func(p *Pipeline) error {
		// Reject absolute paths
		if filepath.IsAbs(dir) {
			return errors.New("absolute paths are not allowed")
		}

		// Prevent navigation to parent directories
		cleanDir := filepath.Clean(dir)
		if cleanDir == ".." || filepath.HasPrefix(cleanDir, "../") {
			return errors.New("parent directory access is not allowed")
		}

		// Join the sanitized relative path with the current directory
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
	}, 

    Exec(func(p *Pipeline) error {
        p.directory = p.mainDirectory
        return nil
    })
}

// SH Executes a command in the directory of the current agent
func SH(name string, args ...string) func(p *Pipeline) error {
    return func(p *Pipeline) error {
        cmd := exec.Command(name, args...)
        cmd.Dir = p.directory
        out, err := cmd.Output()
        log.Println(out)
        return err
    }
}
