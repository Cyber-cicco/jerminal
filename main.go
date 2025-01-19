package main

import (
	. "github.com/Cyber-cicco/jerminal/jerminal"
)

// This is what i want my jerminal application workflow to look like
func main() {

	// Should be an async func
	pipeline := SetPipeline(
		Agent("agent1"),
		RunOnce(
			Exec(SH("git", "clone git@github.com:Cyber-cicco/leeveen-backend.git")),
			Exec(SH("git", "checkout main")),
		),
		Stages("Pull stage",
			Stage(
				"Git Pull",
				Exec(SH("git", "pull")),
			).
                Retry(3, 10),
		),
		Stages("Test Stages",
			Stage(
				"Test controller",
				ExecDefer(CD("internals/controller")),
				Exec(SH("go", "test")),
			),
			Stage(
				"Test services",
				ExecDefer(CD("internals/services")),
				Exec(SH("go", "test")),
			),
		).Parallel(),
	)

}
