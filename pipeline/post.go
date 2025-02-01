package pipeline

type post struct {
	success Success
	failure Failure
	always  Always
}

type Success func(p *Pipeline) error
type Failure func(p *Pipeline) error
type Always func(p *Pipeline) error

func (p *post) ExecuteInPipeline(pipeline *Pipeline) error {
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
