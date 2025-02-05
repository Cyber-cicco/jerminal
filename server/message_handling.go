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
	case "launch-pipeline":
		return s.startPipeline(req, content)

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

func (s *Server) startPipeline(req *rpc.JRPCRequest, content []byte) ([]byte, error) {
    var innerReq rpc.StartPipelineReq
    err := json.Unmarshal(content, &innerReq)
    if err != nil {
        return unMarshalError(req)
    }
    err = s.BeginPipeline(innerReq.Params.Name)
    
    if err != nil {
        return invalidParamsError(req, err)
    }
    res := rpc.JRPCSuccess[rpc.SimpleMessage]{
    	JRPCResponse: rpc.JRPCResponse{
    		RPC: "2.0",
    		ID:  &req.Id,
    	},
    	Value:        rpc.SimpleMessage{
    		Message: fmt.Sprintf("Pipeline %s started successfully", innerReq.Params.Name),
    	},
    }
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

// unMarshalError is an helper function to signify the client that
// his request was not formatted properly
func unMarshalError(req *rpc.JRPCRequest) ([]byte, error) {
	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: "Parmas could not be parsed",
		Data:    nil,
	})
	bytes, err := json.Marshal(res)
	return bytes, err
}

// invalidParamsError is an helper function to signify the client
// has sent invalid data to the server
func invalidParamsError(req *rpc.JRPCRequest, err error) ([]byte, error) {
	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: err.Error(),
		Data:    nil,
	})
	bytes, err := json.Marshal(res)
	return bytes, err
}

// BeginPipeline starts a pipeline cloned from the original one.
// Weird choice, should two of the same pipelines be able to
// execute simultaneously ?
// TODO : figure it out. If we decide to allow it, there will be 
// problems with how agents are handled
// If we accept this, AnyAgent must set the agent of the pipeline
// at runtime
// I thinks it's ok now
func (s *Server) BeginPipeline(id string) error {
	s.store.Lock()
	pipeline, ok := s.store.GlobalPipelines[id]
	s.store.Unlock()
	if !ok {
		fmt.Printf("Wrong id received %s", id)
		return errors.New("Wrong id received")
	}

    // Get a shallow copy of the pipeline
	clone := pipeline.Clone()
	ctx, cancelPipeline := context.WithCancel(context.Background())
	s.activePipelines.Store(clone.GetId(), cancelPipeline)

	go func() {

		defer s.activePipelines.Delete(clone.GetId())
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
    return nil
}
