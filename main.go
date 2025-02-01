package main

import (
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
		Stages("test_stages",
			Stage("test_stage_1",
				SH("touch", "/home/hijokaidan/PC/jerminal/test_2.txt"),
			),
		),
	)
	if err != nil {
		panic(err)
	}
	s := server.New(8002)
    s.SetPipelines([]*Pipeline{p1, p2})
	s.Listen()
}
