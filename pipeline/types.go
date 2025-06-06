package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
)

type EImportance uint8

const (
	DEBUG = EImportance(iota)
	INFO
	WARN
	ERROR
	CRITICAL
)


// MarshalJSON converts DEImp to the corresponding string
func (imp EImportance) MarshalJSON() ([]byte, error) {
    if int(imp) < len(IMPORTANCE_STR) {
        return json.Marshal(IMPORTANCE_STR[imp])
    }
    return json.Marshal(uint8(imp))
}

// UnmarshalJSON converts string back to DEImp
func (imp *EImportance) UnmarshalJSON(data []byte) error {
    // First try to unmarshal as string
    var str string
    if err := json.Unmarshal(data, &str); err == nil {
        for i, s := range IMPORTANCE_STR {
            if s == str {
                *imp = EImportance(i)
                return nil
            }
        }
        return fmt.Errorf("invalid importance string: %s", str)
    }

    return fmt.Errorf("invalid importance value: %s", string(data))
}

// pipelineEvents represents a generic event of the pipeline.
// Each event must be able to execute within a pipeline and provide metadata.
//
// Implemented by : stages, OnceRunner, Post
type pipelineEvents interface {
	ExecuteInPipeline(p *Pipeline, ctx context.Context) error // Executes the component within the pipeline.
	GetShouldStopIfError() bool          // Indicates if the pipeline should stop on error.
	GetName() string
}

// Can be Diagnostic or DiagEvent
type pipelineLog interface {
    Log()
}

// executable represents an entity that can be executed within a pipeline.
//
// Implemented by Exec, executor
type executable interface {
	Execute(p *Pipeline, ctx context.Context) error // Executes the entity.
}
