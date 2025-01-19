package jerminal

import (
	"log"
	"os/exec"
)

func SH(name string, args ...string) func(p *Pipeline) error {
    return func(p *Pipeline) error {
        cmd := exec.Command(name, args...)
        cmd.Dir = p.Directory
        out, err := cmd.Output()
        log.Println(out)
        return err
    }
}
