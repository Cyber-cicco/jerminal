package main

import (
	"errors"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {
	i := 0
	standardPost := Post(
		Success(func(p *Pipeline) error {
			p.Diagnostic.NewDE(INFO, "Job was successfull")
			return nil
		}),
		Failure(func(p *Pipeline) error {
			p.Diagnostic.NewDE(INFO, "Job failed")
			return nil
		}),
		Always(func(p *Pipeline) error {
			return nil
		}),
	)
	p1, err := SetPipeline("test1",
		AnyAgent(),
		Stages("test_stages",
			Stage("test_stage_1",
				SH("touch", "/home/hijokaidan/PC/jerminal/test_1.txt"),
			),
			Stage("test",
				Exec(func(p *Pipeline) error {
					if i < 2 {
						i++
						return errors.New("test error")
					}
					return nil
				}),
			).Retry(2, 1),
		),
		standardPost,
	)
	p1.ReportJson()
	p2, err := SetPipeline("test2",
		Agent("agent_2"),
		RunOnce(
			SH("git", "clone", "git@github.com:Cyber-cicco/symfgoni.git"),
		),
		Stages("symfgoni build",
			Stage("get_latest_version",
				CD("symfgoni"),
				SH("git", "pull"),
			),
			Stage("test",
				CD("symfgoni/"),
				SH("go", "test", "./..."),
			),
			Stage("build",
				CD("symfgoni/internals"),
				SH("go", "build", "-o", "exe"),
				SH("cp", "exe", "/home/hijokaidan/PC/jerminal/exe"),
			),
		),
		standardPost,
	)
	p3, err := SetPipeline("test3", // 1 diag event for the start
		AnyAgent(),
		Stages("stages_1", // 1 diag for the stages
			Stage("stage_1", // 1 diag for the stage
				SH("echo", "bonjour"), // 1 diag event
				SH("echo", "bonjour"), // 1 diag event
				SH("echo", "bonjour"), // 1 diag event
			), // 1 at the end of stages_1
			Stage("stage_2", // 1 diag for the stage
				SH("echo", "bonjour"), // 1 diag event
				SH("echo", "bonjour"), // 1 diag event
				SH("echo", "bonjour"), // 1 diag event
			), // 1 at the end of stages_2
		), // 1 diag at the end
	)
	p3.ReportJson()
	if err != nil {
		panic(err)
	}
	s := server.New(8002)
	s.SetPipelines([]*Pipeline{p1, p2, p3})
	s.Listen()
}
