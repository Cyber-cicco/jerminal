package main

import (
	"context"
	"errors"
	"time"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {
	i := 0
	standardPost := Post(
		Success(func(p *Pipeline, ctx context.Context) error {
			p.Diagnostic.NewDE(INFO, "Job was successfull")
			return nil
		}),
		Failure(func(p *Pipeline, ctx context.Context) error {
			p.Diagnostic.NewDE(INFO, "Job failed")
			return nil
		}),
		Always(func(p *Pipeline, ctx context.Context) error {
			return nil
		}),
	)
	p1, err := SetPipeline("test1",
		AnyAgent(),
		RunOnce(
			SH("touch", "mytralala"),
		),
		Stages("test_stages",
			Stage("test_stage_1",
				SH("touch", "/home/hijokaidan/PC/jerminal/test_1.txt"),
			),
			Stage("test",
				Exec(func(p *Pipeline, ctx context.Context) error {
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
	p2.ReportJson()
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
	p4, err := SetPipeline("test4",
		AnyAgent(),
		Stages("stages_1",
			Stage("stage_1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
			),
		),
		Stages("stages_2",
			Stage("stage_1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
			),
		),
		Stages("stages_3",
			Stage("stage_1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
			),
		),
		Stages("stages_4",
			Stage("stage_1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					time.Sleep(time.Second * 5)
					return nil
				}),
			),
		),
	)
	p4.ReportJson()
	if err != nil {
		panic(err)
	}
	s := server.New(8002)
	s.SetPipelines([]*Pipeline{p1, p2, p3, p4})
	s.Listen()
}
