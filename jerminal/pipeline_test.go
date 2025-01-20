package jerminal

import (
	"testing"

	"github.com/Cyber-cicco/jerminal/jerminal/state"
)

func TestPipelineExecution(t *testing.T) {
	p := setPipelineWithState("test",
		Agent(""),
        state.GetStateCustomConf(&state.Config{}),
	)
    t.Log(p)
}
