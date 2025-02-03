package pipeline

import (
	"context"
)

type post struct {
	success Success
	failure Failure
	always  Always
}

// Invoked by post in case of success
type Success func(p *Pipeline) error
// Invoked by post in case of failure
type Failure func(p *Pipeline) error
// Always invoked by post
type Always func(p *Pipeline) error

func (p *post) ExecuteInPipeline(pipeline *Pipeline, ctx context.Context) error {
    diag := NewDiag("post")
    pipeline.Diagnostic.AddChild(diag)
    pipeline.Diagnostic = diag
    defer func(){
        pipeline.ResetDiag()
    }()
    if pipeline.Inerror {
        err := p.failure.ExecuteError(pipeline)
        if err != nil {
            return err
        }
    } else {
        err := p.success.ExecuteSuccess(pipeline)
        if err != nil {
            return err
        }
    }
    return p.always.ExecuteAlways(pipeline)
}

func (p *post) GetName() string {
    return "Post pipeline job"
}

// Functions to execute after executing stages.
// Technically, you could use Post anywhere in the
// pipeline, but it is not recommended
func Post(success Success, failure Failure, always Always) *post {
    return &post{
    	success: success,
    	failure: failure,
    	always:  always,
    }
}

// GetShouldStopIfError must always return true for this struct
func (p *post) GetShouldStopIfError() bool {
	return true
}

func (s Success) ExecuteSuccess(p *Pipeline) error {
    return s(p)
}

func (f Failure) ExecuteError(p *Pipeline) error {
    return f(p)
}

func (a Always) ExecuteAlways(p *Pipeline) error {
    return a(p)
}
