package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

const (
	PARSE_ERROR      = -32700
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603
)

// Received structure to decode in JSON
type JRPCRequest struct {
	JsonRpcVersion string `json:"jsonprc"`
	Id             int    `json:"id"`
	Method         string `json:"method"`

	//Param
}

type CustomJRPCRequest struct {
	JRPCRequest
	Params map[string]any
}

// Response sent by the server to the client after receiving
// a JRPCRequest
type JRPCResponse struct {
	RPC string `json:"jsonrpc"`
	ID  *int   `json:"id"`

	//Result | error
}

// Error type for JRPC Response
type JRPCError struct {
	JRPCResponse
	Error ErrorData `json:"error"`
}

type JRPCSuccess[T any] struct {
	JRPCResponse
	Value T `json:"value"`
}

// ErrorData is the Response interface for whenever it encounters an error
type ErrorData struct {
	Code    int16       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CancelationReqParams struct {
	PipelineId             string `json:"pipeline-id"`        // Unique identifier of the pipeline to cancel
	PipeLineLifetimeSecret string `json:"pipeline-lt-secret"` // Secret ensuring the process has the rights to perform cancelation
}

type CancelationReq struct {
	JRPCRequest
	Params CancelationReqParams `json:"params"`
}

// ListPipelinesParams list request options to get
// a json representation of a / multiple pipelines
type ListPipelinesParams struct {
	Id     *string `json:"id,omitempty"` // Id if the pipeline to list. If not present, return every pipeline
	Active bool    `json:"active"`       // To know if the pipeline to search is a running process
	All    bool    `json:"all"`          // tells if we need to see every pipeline
}

type GetPipelinesReq struct {
	JRPCRequest
	Params ListPipelinesParams `json:"params"`
}

type StartPipelineReq struct {
	JRPCRequest
	Params StartPipelineParams `json:"params"`
}

type StartPipelineParams struct {
	Name string
}

type GetReportsReq struct {
	JRPCRequest
	Params GetReportsParams
}

type GetReportsParams struct {
	PipelineId    *string  `json:"pipeline-id"`    // To specify if you want the reports of a single run of a pipeline
	PipelineName  *string  `json:"pipeline-name"`  // To specify if you want all the reports of a named pipeline
	Type          string   `json:"type"`           // Specify the report type
	Fields        []string `json:"fields"`         // To specify if you only want some fields of the pipeline.
	OmittedFields []string `json:"omitted-fields"` // To specify if you want every info except one
}

type SimpleMessage struct {
	Message string `json:"message"`
}

type TerminationReq struct {
}

// SplitFunc allows to use a scanner to parse the JRPCRequest
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	header, content, found := bytes.Cut(data, []byte{'\r', '\n', '\r', '\n'})

	if !found {
		return 0, nil, nil
	}

	contentLenBytes := header[16:]
	contentLength, err := strconv.Atoi(string(contentLenBytes))

	if err != nil {
		return 0, nil, err
	}

	if len(content) < contentLength {
		return 0, nil, nil
	}

	totalLength := len(header) + 4 + contentLength
	return totalLength, data[:totalLength], nil
}

// DecodeMessage gets the content of the message as deserializable JSON, and the method type
// from a standard PMSP client message.
func DecodeMessage[T any](msg []byte) (*T, []byte, error) {
	header, content, found := bytes.Cut(msg, []byte{'\r', '\n', '\r', '\n'})

	if !found {
		return nil, nil, errors.New("Message header not found")
	}

	contentLenBytes := header[16:]
	contentLength, err := strconv.Atoi(string(contentLenBytes))

	if err != nil {
		return nil, nil, err
	}

	var baseMessage T

	err = json.Unmarshal(content[:contentLength], &baseMessage)
	if err != nil {
		return nil, nil, err
	}
	return &baseMessage, content[:contentLength], nil
}

// NewError helps build a standardized JRPC error response
func NewError(reqId *int, err ErrorData) JRPCError {

	return JRPCError{
		JRPCResponse: JRPCResponse{
			RPC: "2.0",
			ID:  reqId,
		},
		Error: err,
	}
}

// NewResult helps build a standardized JRPC success response
func NewResult[T any](reqId int, value T) JRPCSuccess[T] {
	return JRPCSuccess[T]{
		JRPCResponse: JRPCResponse{
			RPC: "2.0",
			ID:  &reqId,
		},
		Value: value,
	}
}

// JRPCRes allows
func JRPCRes(bytes []byte) []byte {
	res := []byte("Content-Length: ")
	length := len(bytes)
	res = append(res, []byte(strconv.Itoa(length)+"\r\n\r\n")...)
	res = append(res, bytes...)
	return res
}
