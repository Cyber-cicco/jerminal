package main

import (
	"fmt"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server"
)

func main() {
	p1, err := SetPipeline("test1",
		AnyAgent(),
		Stages("test_stages",
			Stage("test_stage_1",
				SH("touch", "/home/hijokaidan/PC/jerminal/test_1.txt"),
			),
		),
	)
	p2, err := SetPipeline("test2",
		Agent("agent_2"),
		RunOnce(
			SH("git", "clone", "git@github.com:Cyber-cicco/symfgoni.git"),
		),
		Stages("get",
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
		Post(
			Success(func(p *Pipeline) error {
                fmt.Printf("Job was successfully executed")
                return nil
            }),
            Failure(func(p *Pipeline) error {
                fmt.Println("Job failed")
                return nil
            }),
            Always(func(p *Pipeline) error {
                for _, ev := range p.Diagnostic.Events {
                    fmt.Printf("%v, %v\n", ev.Importance, ev.Description)
                }
                return nil
            }),
		),
	)
	if err != nil {
		panic(err)
	}
	s := server.New(8002)
	s.SetPipelines([]*Pipeline{p1, p2})
	s.Listen()
}
