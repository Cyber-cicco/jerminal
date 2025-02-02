package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

// Received structure to decode in JSON
type JRPCRequest struct {
	JsonRpcVersion string `json:"jsonprc"`
	Id             int    `json:"id"`
	Method         string `json:"method"`

	//Param
}

type JRPCResponse struct {

	//Param
}

type CancelationRequest struct {
	JRPCRequest
	Params CancelationRequestParams
}

type CancelationRequestParams struct {
	PipelineId                     string // Unique identifier of the pipeline to cancel
	PipeLineLifetimeSecret string // Secret ensuring the process has the rights to perform cancelation
}

type TerminationRequest struct {
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
func DecodeMessage(msg []byte) (string, []byte, error) {
	header, content, found := bytes.Cut(msg, []byte{'\r', '\n', '\r', '\n'})

	if !found {
		return "", nil, errors.New("Message header not found")
	}

	contentLenBytes := header[16:]
	contentLength, err := strconv.Atoi(string(contentLenBytes))

	if err != nil {
		return "", nil, err
	}

	var baseMessage JRPCRequest

	err = json.Unmarshal(content[:contentLength], &baseMessage)
	if err != nil {
		return "", nil, err
	}
	return baseMessage.Method, content[:contentLength], nil
}
