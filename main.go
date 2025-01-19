package main

import (
	. "github.com/Cyber-cicco/jerminal/jerminal"
)

// This is what i want my jerminal application workflow to look like
func main() {

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
			).DontStopIfErr(),
			Stage(
				"Test services",
				ExecDefer(CD("internals/services")),
				Exec(SH("go", "test")),
			).DontStopIfErr(),
		).Parallel(),
		Stages("Build Stages",
			Stage("Docker Build",
				ExecTryCatch(
					SH("docker", "build leeveen-backend"),
					Stage("Docker remove",
						Exec(SH("docker", "stop leeveen-backend")),
						Exec(SH("docker", "remove leeveen-backend")),
					),
				),
			).Retry(1, 1),
			Stage("Docker deploy",
				ExecTryCatch(
					SH("docker", "compose up"),
					Exec(SH("docker", "compose down")),
				),
			).Retry(1, 1),
		),
	)
}
