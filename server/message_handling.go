package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
)

// handleMessage checks for the message type and calls the appropriate function
func (s *Server) handleMessage(req *rpc.JRPCRequest, content []byte) ([]byte, error) {
	switch req.Method {

	case "pipeline-cancelation":
		return s.cancelPipeline(req, content)

	case "list-existing-pipelines":
		return s.listExistingPipelines(req, content)

	default:
		res := rpc.NewError(&req.Id, rpc.ErrorData{
			Code:    rpc.METHOD_NOT_FOUND,
			Message: fmt.Sprintf("Method %s is not supported", req.Method),
		})
		bytes, err := json.Marshal(res)
		return bytes, err
	}
}

func (s *Server) cancelPipeline(req *rpc.JRPCRequest, content []byte) ([]byte, error) {
	var cancelParams rpc.CancelationReq
	err := json.Unmarshal(content, &cancelParams)
	if err != nil {
		return unMarshalError(req)
	}
	err = s.cancelPipelineByLabel(cancelParams)
	if err != nil {
        return invalidParamsError(req, err)
	}
	res := rpc.NewResult(req.Id, "cancelation succeeded")
	return json.Marshal(res)
}

// listExistingPipelines gets back one pipeline based on it's id, or, if the 
// id is not present, a set of pipelines
func (s *Server) listExistingPipelines(req *rpc.JRPCRequest, content []byte) ([]byte, error) {

    s.store.Lock()
    defer s.store.Unlock()

	var params rpc.GetPipelinesReq
	err := json.Unmarshal(content, &params)
	if err != nil {
		return unMarshalError(req)
	}

    pipelineMap := s.getMapBasedOnActive(params.Params.Active) 

	if params.Params.Id != nil {
        var res rpc.JRPCSuccess[*pipeline.Pipeline]
        pipeline, ok := pipelineMap[*params.Params.Id]
        if !ok {
            return invalidParamsError(req, errors.New("Pipeline not found"))
        }

        res.Value = pipeline
        return json.Marshal(res)
	}

    pipelines := make([]*pipeline.Pipeline, len(pipelineMap))
    i := 0
    for _, v := range pipelineMap {
        pipelines[i] = v
        i++
    }

    return json.Marshal(pipelines)
}

func (s *Server) getMapBasedOnActive(active bool) map[string]*pipeline.Pipeline {
    if active {
        return s.store.ActivePipelines
    }
    return s.store.GlobalPipelines
}

// Cancel a specific pipeline by its label
func (s *Server) cancelPipelineByLabel(cancelParams rpc.CancelationReq) error {
	fmt.Println("Cancelling the pipeline")
	fn, ok := s.activePipelines.Load(cancelParams.Params.PipelineId)
	if !ok {
		return errors.New("Pipeline not found")
	}

	cancelFunc := fn.(context.CancelFunc)
	cancelFunc()
	return nil
}

func unMarshalError(req *rpc.JRPCRequest) ([]byte, error) {

	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: "Parmas could not be parsed",
		Data:    nil,
	})
	bytes, err := json.Marshal(res)
	return bytes, err
}

func invalidParamsError(req *rpc.JRPCRequest, err error) ([]byte, error) {
	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: err.Error(),
		Data:    nil,
	})
	bytes, err := json.Marshal(res)
	return bytes, err
}

// BeginPipeline starts a pipeline from the Id of ti
func (s *Server) BeginPipeline(id string) {
    s.store.Lock()
	pipeline, ok := s.store.GlobalPipelines[id]
    s.store.Unlock()
	if !ok {
		fmt.Printf("Wrong id received %s", id)
		return
	}
	clone := pipeline.Clone()

	// Create a new context for this pipeline execution
	ctx, cancelPipeline := context.WithCancel(context.Background())

	executionID := clone.GetId()

	// Store the cancel function
	s.activePipelines.Store(executionID, cancelPipeline)

	// Create channel for cleanup coordination
	done := make(chan struct{})

	go func() {

		defer close(done)
		defer s.activePipelines.Delete(executionID)
		defer cancelPipeline()

        s.store.Lock()
        s.store.ActivePipelines[clone.GetId()] = &clone
        s.store.Unlock()

		err := clone.ExecutePipeline(ctx)
		if err != nil {
            s.store.Lock()
            defer s.store.Unlock()
			if err == context.Canceled {
				fmt.Printf("Pipeline '%s' was cancelled\n", pipeline.Name)
			} else {
				fmt.Printf("Pipeline '%s' failed with error: %v\n", pipeline.Name, err)
			}
            delete(s.store.ActivePipelines, clone.GetId())
		}
	}()

	// Wait for either context cancellation or pipeline completion
	select {
	case <-ctx.Done():
		fmt.Printf("Pipeline '%s' cancelled\n", pipeline.Name)
	case <-done:
		fmt.Printf("Pipeline '%s' completed\n", pipeline.Name)
	}
}

