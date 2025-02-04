package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	var cancelParams rpc.CancelationRequest
	err := json.Unmarshal(content, &cancelParams)
	if err != nil {
		return unMarshalError(req)
	}
	err = s.cancelPipelineByLabel(cancelParams)
	if err != nil {
        return invalidParamsError(req, err)
	}
	res := rpc.NewResult(req.Id, "cancelation succeeded")
	bytes, err := json.Marshal(res)
	return bytes, err
}

func (s *Server) listExistingPipelines(req *rpc.JRPCRequest, content []byte) ([]byte, error) {
	var params rpc.ListPipelinesRequest
	err := json.Unmarshal(content, &params)
	if err != nil {
		return unMarshalError(req)
	}
	if params.Params.Id != nil {
		pipelines := s.pipelines
        fmt.Printf("pipelines: %v\n", pipelines)
        //TODO : changer
        return nil, nil
	}
    return nil, nil
}

// Cancel a specific pipeline by its label
func (s *Server) cancelPipelineByLabel(cancelParams rpc.CancelationRequest) error {
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
