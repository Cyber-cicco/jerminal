package jerminal

import "testing"

func TestPipelineExecution(t *testing.T) {
	p := SetPipeline("test",
		Agent(""),
	)
    t.Log(p)
}
