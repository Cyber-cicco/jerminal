package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Cyber-cicco/jerminal/utils"
)

const version = "0.1.1"

var (
    generateFlag string
    resourcesDir = os.ExpandEnv(fmt.Sprintf("${GO_PATH}/pkg/mod/github.com/cyber-cicco/jerminal@%s/resources/", version))
)

func init() {
    flag.StringVar(&generateFlag, "gen", "", "Used to generate jerminal project files. Argument is the name of your go module")

}

func main() {
    flag.Usage = usage
    flag.Parse()

    if generateFlag == "" {
        fmt.Println("generate flag is required")
        flag.Usage()
        os.Exit(1)
    }

    sh("go", "mod", "init", generateFlag)
    sh("go", "get", "github.com/Cyber-cicco/jerminal")
    sh("go", "mod", "tidy")

    err := utils.CopyDir(resourcesDir, "./resources")
    if err != nil {
        panic(err)
    }

}

func sh(name string, args ...string) {

    cmd := exec.Command(name, args...)
    _, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Got error creating go module %v", err)
        panic(err)
    }
}

func usage() {
    fmt.Println("Jerminal Generator", version)
    fmt.Println()
    fmt.Println("Usage : ")
    order := []string{"gen"}

    for _, name := range order {
        flagEntry := flag.CommandLine.Lookup(name)
		fmt.Printf("  -%s\n", flagEntry.Name)
		fmt.Printf("\t%s\n", flagEntry.Usage)

    }

    fmt.Println(`Example commands:
    $ jerminal -gen github.com/Cyber-cicco/custom-ci
    `)

}
