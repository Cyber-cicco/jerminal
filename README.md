# Jerminal: Jenkins in the Terminal

⚠️ This is a work in progress. 

## What is Jerminal?

Jerminal is a pipeline framework that allows you to describe a pipeline of events you want
to execute with helper functions. It provides:

- Detailed event logs
- A system of agents and schedules with fine-grained customization
- A unix socket server
- GitHub webhooks integration
- Report generation in JSON files (MongoDB and SQLite support coming soon)

You can use Jerminal for:
* CI/CD pipelines
* Deterministic simulation testing
* Integration testing
* Load balancing
* Scheduled task execution

## Why Jerminal?

After 2 days of using Jenkins and wanting to `rm -rf / --no-preserve-root` myself IRL, I concluded that a web interface isn't always the best approach for pipeline management. Jerminal's philosophy is simple: pipeline configuration and execution are more efficient when done through code in a terminal environment.

Benefits of Jerminal over traditional CI/CD tools:
- Configure pipelines with code in your IDE
- Use a type-safe scripting language
- Version control your pipeline configurations
- Easier tool and dependency management
- No web interface complexity

## Installation

```bash
go get github.com/Cyber-cicco/jerminal
```

## Getting Started

### Basic Pipeline

You can define a pipeline using the `SetPipeline` function:

```go
import (
	"errors"
	"fmt"
	"context"

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
	if err != nil {
		// Handle error
	}
	
	// Start the pipeline
	pipeline.Start(context.Background())
}
```

### Pipeline Structure

A pipeline consists of:

1. **Name**: Defined by the first argument of `SetPipeline`
2. **Agent**: Gets its own directory to execute code from. Files created by the agent (not cached) are removed when execution completes
3. **Stages**: Executed sequentially by default

Each `Stages` object takes a set of `Stage` objects that contain functions to execute.

### Advanced Configuration

- **Parallel Execution**: Configure stages to run in parallel
- **Retry Logic**: Set stages to retry a specified number of times with delay
- **Deferred Functions**: Run cleanup code at the end of stages
- **Parameter Passing**: Pass data between pipeline stages

### Setting Up a Server

Create a server to manage your pipelines:

```go
import (
	"errors"
	"context"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {
    // Pipeline definition...

	s := server.New()
	s.SetPipelines(pipeline)
	s.ListenGithubHooks(8091)
}
```

- `server.New()` sets up a unix socket server that listens for JSON RPC messages
- `s.ListenGithubHooks(8091)` listens for GitHub webhooks on the specified port

## Command Reference

### Key Functions

- `SetPipeline(name, agent, ...commands)`: Create a new pipeline
- `AnyAgent()`: Create a generic agent for execution
- `RunOnce(...)`: Execute commands only on first run
- `Stages(name, ...stages)`: Group stages together
- `Stage(name, ...commands)`: Define an execution stage
- `SH(command, ...args)`: Execute shell commands
- `Exec(func)`: Run custom Go functions

### Stage Modifiers

- `.Retry(attempts, delay)`: Configure retry behavior
- `.Parallel()`: Run stages in parallel
- `.Defer(func)`: Execute after stage completion

## Configuration

Jerminal uses JSON configuration files located in the `resources` directory:

- `jerminal.json`: Core application settings
- `agents.json`: Agent configuration

## Examples

Check the `integration_tests` directory for complete examples:

- Pipeline creation and execution
- GitHub webhook integration
- Report generation
- Command execution

## Contributing

Contributions are welcome! Please see the wiki for development guidelines.

## License

MIT
