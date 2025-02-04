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
        return json.Marshal(pipeline)
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
