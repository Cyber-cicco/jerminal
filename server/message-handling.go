package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
	"github.com/Cyber-cicco/jerminal/utils"
)

// handleMessage checks for the message type and calls the appropriate function
func (s *Server) handleMessage(req *rpc.JRPCRequest, content []byte) []byte {
	switch req.Method {

	case "pipeline-cancelation":
		return s.cancelPipeline(req, content)
	case "list-existing-pipelines":
		return s.listExistingPipelines(req, content)
	case "launch-pipeline":
		return s.startPipeline(req, content)
	case "get-reports":
		return s.getReports(req, content)

	default:
		res := rpc.NewError(&req.Id, rpc.ErrorData{
			Code:    rpc.METHOD_NOT_FOUND,
			Message: fmt.Sprintf("Method %s is not supported", req.Method),
		})
		return utils.MustMarshall(res)
	}
}

// startPipeline cancels a running pipeline based on it's id
//
// Handle all errors internally and sends back the appropriate response
// without the caller needing to handle the state.
func (s *Server) cancelPipeline(req *rpc.JRPCRequest, content []byte) []byte {
	var cancelParams rpc.CancelationReq
	err := json.Unmarshal(content, &cancelParams)
	if err != nil {
		return paramsError(req)
	}
	err = s.cancelPipelineByLabel(cancelParams)
	if err != nil {
		return invalidParamsError(req, err)
	}
	res := rpc.NewResult(req.Id, "cancelation succeeded")
	return utils.MustMarshall(res)
}

// startPipeline starts a pipeline based on the name provided in the rpc
// request.
//
// Handle all errors internally and sends back the appropriate response
// without the caller needing to handle the state.
func (s *Server) startPipeline(req *rpc.JRPCRequest, content []byte) []byte {
	var innerReq rpc.StartPipelineReq
	err := json.Unmarshal(content, &innerReq)
	if err != nil {
		return paramsError(req)
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
		Value: rpc.SimpleMessage{
			Message: fmt.Sprintf("Pipeline %s started successfully", innerReq.Params.Name),
		},
	}
	return utils.MustMarshall(res)
}

// listExistingPipelines gets back one pipeline based on it's id, or, if the
// id is not present, a set of pipelines
func (s *Server) listExistingPipelines(req *rpc.JRPCRequest, content []byte) []byte {

	s.store.Lock()
	defer s.store.Unlock()

	var params rpc.GetPipelinesReq
	err := json.Unmarshal(content, &params)
	if err != nil {
		return paramsError(req)
	}

	pipelineMap := s.getMapBasedOnActive(params.Params.Active)

	if params.Params.Id != nil {
		var res rpc.JRPCSuccess[*pipeline.Pipeline]
		pipeline, ok := pipelineMap[*params.Params.Id]
		if !ok {
			return invalidParamsError(req, errors.New("Pipeline not found"))
		}

		res.Value = pipeline
		return utils.MustMarshall(res)
	}

	if params.Params.All {
		pipelines := make([]*pipeline.Pipeline, len(pipelineMap))
		i := 0
		for _, v := range pipelineMap {
			pipelines[i] = v
			i++
		}

		return utils.MustMarshall(pipelines)
	}
	return invalidParamsError(req, errors.New("Invalid format for pipeline start request"))

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
	if cancelParams.Params.PipelineId == "" {
		return errors.New("ID must be given")
	}
	fn, ok := s.activePipelines.Load(cancelParams.Params.PipelineId)
	if !ok {
		return errors.New("Pipeline not found")
	}

	cancelFunc := fn.(context.CancelFunc)
	cancelFunc()
	return nil
}

// paramsError is an helper function to signify the client that
// his request was not formatted properly
func paramsError(req *rpc.JRPCRequest) []byte {
	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: "Params could not be parsed",
		Data:    nil,
	})
	return utils.MustMarshall(res)
}

// marshallError is an helper function to signify the client
// that the body of the request is invalid JSON
func marshallError() []byte {
	res := rpc.NewError(nil, rpc.ErrorData{
		Code:    rpc.PARSE_ERROR,
		Message: "Client sent invalid JSON",
		Data:    nil,
	})
	return utils.MustMarshall(res)
}

// invalidParamsError is an helper function to signify the client
// has sent invalid data to the server
func invalidParamsError(req *rpc.JRPCRequest, err error) []byte {
	res := rpc.NewError(&req.Id, rpc.ErrorData{
		Code:    rpc.INVALID_PARAMS,
		Message: err.Error(),
		Data:    nil,
	})
	return utils.MustMarshall(res)
}

// BeginPipeline starts a pipeline cloned from the original one.
func (s *Server) BeginPipeline(id string) error {
	s.store.Lock()
	pipeline, ok := s.store.GlobalPipelines[id]
	s.store.Unlock()
	if !ok {
		return fmt.Errorf("Wrong id received %s", id)
	}

	// Get a shallow copy of the pipeline
	clone := pipeline.Clone()
	ctx, cancelPipeline := context.WithCancel(context.Background())
	s.activePipelines.Store(clone.GetId(), cancelPipeline)

	go func() {

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
		s.store.Lock()
		delete(s.store.ActivePipelines, clone.GetId())
		s.store.Unlock()
	}()
	return nil
}

func (s *Server) getReports(req *rpc.JRPCRequest, content []byte) []byte {
	var params rpc.GetReportsReq
	err := json.Unmarshal(content, &params)
	if err != nil {
		return paramsError(req)
	}

	if params.Params.PipelineName == nil {
		return paramsError(req)
	}

	switch params.Params.Type {
	case "json":
		return s.getReportsFromJson(&params)
	default:
		return paramsError(req)
	}
}

// getReportsFromJson handles the GetReportsReq to get back reports
// of executions of a pipeline
func (s *Server) getReportsFromJson(req *rpc.GetReportsReq) []byte {
	params := req.Params
	dir := filepath.Join(s.config.ReportDir, *req.Params.PipelineName)

	if params.PipelineId != nil {
		return s.getReportFromId(req, *req.Params.PipelineId, dir)
	}
    return s.getAllReports(req, dir)

}

// getAllReports finds all the reports in the report directory of the pipeline
func (s *Server) getAllReports(req *rpc.GetReportsReq, dir string) []byte {
	maps := []map[string]interface{}{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Handle any walk errors
		if err != nil {
            fmt.Printf("err: %v\n", err)
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		s.getReportFromId(req, info.Name(), dir)
		content, err := unMarshallFileFromReq(req, dir, strings.TrimSuffix(info.Name(), ".json") )
		if err != nil {
            fmt.Printf("err: %v\n", err)
			return err
		}
		maps = append(maps, content)
		return nil
	})

	if err != nil {
		err := rpc.NewError(&req.Id, rpc.ErrorData{
			Code:    rpc.INVALID_PARAMS,
			Message: "File could not be found or is in invalid format",
		})
		return utils.MustMarshall(err)
	}

	return utils.MustMarshall(maps)
}

// getReportFromId gets back a report from the id provided in the request
func (s *Server) getReportFromId(req *rpc.GetReportsReq, id, dir string) []byte {
	content, err := unMarshallFileFromReq(req, id, dir)
	if err != nil {
		err := rpc.NewError(&req.Id, rpc.ErrorData{
			Code:    rpc.INVALID_PARAMS,
			Message: "File could not be found or is in invalid format",
		})
		return utils.MustMarshall(err)
	}
	res := rpc.NewResult(req.Id, content)
	return utils.MustMarshall(res)
}
