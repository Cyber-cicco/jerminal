package main

import (

	. "github.com/Cyber-cicco/jerminal/jerminal"
)

// This is what i want my jerminal application workflow to look like
func main() {

	// Should be an async func
	pipeline := SetPipeline(
		GetAgent("agent1"),
		RunOnce(
			Exec(
				SH("git", "remote add origin git@github.com:Cyber-cicco/leeveen-backend.git"),
			),
			Exec(SH("git", "checkout main")),
		),
		SetStages(
			SetStage(
				"Git gud",
				Exec(SH("git", "pull")),
			),
			SetStage(
				"build project",
				Exec(SH("go", "build")),
			),
		),
	)

}
