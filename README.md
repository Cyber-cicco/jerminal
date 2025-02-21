# Jerminal : Jenkins in the terminal

⚠️  This is a work in progress. 

## What is Jerminal ?

Jerminal is a pipeline framework that allow you to describe a pipeline of events you want
to execute with helper functions. It also provides a set of detailed event logs, and
a system of agents and schedules with fine grained customization.

But more than this, it also provides a set of out of the box functionnalities, like
a unix socket server, an integration with github hooks, and a production of reports in JSON files,
and soon MongoDB and SQLITE.

You can use it for all sorts of things: 
 * CI/CD
 * Deterministic simulation testing
 * Integration testing
 * Load balancing


## Why Jerminal ?

After an intense 2 days of using Jenkins and wanting to `rm -rf / --no-preserve-root` myself IRL,
I came to the conclusion that a web interface was probably not the best way to do what Jenkins
is doing.

The idea behind Jerminal is simple : there is not a single thing in jenkins that wouldn't be easier
if it was done through code and a terminal. Installing a tool, a plugin, scripting the pipeline, all
of this would be so much more easier in an IDE with config files and a scripting language that enables
type safety. That what jerminal does.

## Getting started

You can define a pipeline using the `SetPipeline` function, as such:

```go

import (
	"errors"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {
    i := 0
	pipeline, err := SetPipeline("test1",
		AnyAgent(),
		RunOnce(
			Exec(func(p *Pipeline, ctx context.Context) error {
                fmt.Println("I'm only ran the first time the pipeline gets executed")
                p.Diagnostic.LogEvent(INFO, "This is how you can log an event in the main pipeline")

                return nil
            }),
		),
		Stages("test_stages",
			Stage("test_stage_1",
				SH("touch", "mytralala"),
			),
			Stage("test_stage_2",
				Exec(func(p *Pipeline, ctx context.Context) error {
					if i < 2 {
						i++
						return errors.New("test error")
					}
					return nil
				}),
			).Retry(2, 1),
		),
	)

}
```

A pipeline consist of :

 * a name, defined by the first argument of `SetPipeline`
 * an agent, that gets it's own directory to execute code from. Every file created by the agent that is not cached will be destroyed at the end of the pipeline execution
 * a set of stages, that will be executed sequentially.

Each `stages` object will take a set of `stage` object to be executed, that each contains a set of functions.

There are multiple ways of configuring theses objects. Each stage in a `stages` can be configured to be ran in parallel instead of sequentially, each `stage` can be configured to get retried n number of times, you can defer a function to be executed at the end of a stage, etc. Go check the wiki to find them all.

After configuring your pipeline, you can create a server by using the function `New()` from the package server, in this way :

```go

import (
	"errors"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {

    // Pipeline code goes here...

	s := server.New()
	s.SetPipelines(p1)
	s.ListenGithubHooks(8091)
}

```

`server.new()` will set up a unix socket server that listens for JSON RPC messages and executes pipelines based on what it gets from a client.

`s.ListenGithubHooks(8091)` will listen for github hooks on the specified port of your server, and execute a pipeline based on the last part of the url.

You will be able to find further infos in the wiki

