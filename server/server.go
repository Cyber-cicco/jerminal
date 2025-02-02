package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
)

// Test variable. Should be replaced with config
const TEST_ENV_VAR = "GITHUB_WEBHOOK_SECRET"

// HookServer receives the webhook call and executes pipelines
type HookServer struct {
	pipelines       map[string]*pipeline.Pipeline // map of names to a pipeline
	port            uint16                        // port to listen to
	listener        net.Listener                  // Unix socket listener
	activePipelines sync.Map                      // map[string]context.CancelFunc
}

// New creates a new server to Listen for incoming webhooks
func New(port uint16) *HookServer {
	server := &HookServer{
		pipelines: map[string]*pipeline.Pipeline{},
		port:      port,
	}

	// Set up Unix socket
	socketPath := "/tmp/pipeline-control.sock"
	os.Remove(socketPath) // Clean up existing socket if any

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Failed to create Unix socket: %v\n", err)
		return server
	}

	server.listener = listener

	// Start socket listener in goroutine
	go server.listenForCancellation()

	return server
}

// listenForCancellation waits for a Socket with a message as such:
// cancel <pipeline-name>
//
// Might want to structure this a bit more so you can listen to more
// types of messages
//
// # The structure of the message should be of JSON RPC, like for the LSPs
//
// This would give an interface for other local programs to interact with the process.
func (s *HookServer) listenForCancellation() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			scanner := bufio.NewScanner(c)
			scanner.Split(rpc.SplitFunc)
			for scanner.Scan() {
				msg := scanner.Bytes()
				method, content, err := rpc.DecodeMessage(msg)
				if err != nil {
					fmt.Printf("Error encountered : %s", err)
					continue
				}
				s.handleMessage(method, content)
			}

		}(conn)
	}
}

// Trigger function 
func (s *HookServer) handleMessage(method string, content []byte) error {
	switch method {

	case "pipeline-cancelation":

		var cancelParams rpc.CancelationRequest
		err := json.Unmarshal(content, &cancelParams)
		if err != nil {
			return err
		}
        err = s.cancelPipelineByLabel(cancelParams)
		return err

	default:
		return errors.New("Unsupported method")
	}
}

// Cancel a specific pipeline by its label
func (s *HookServer) cancelPipelineByLabel(cancelParams rpc.CancelationRequest) error {
	s.activePipelines.Range(func(key, value interface{}) bool {
		if cancelParams.Params.PipelineId == key.(string) {
			if cancel, ok := value.(context.CancelFunc); ok {
				cancel()
				s.activePipelines.Delete(key)
			}
		}
		return true
	})
	return nil
}


// Puts the pipelines in the server
func (s *HookServer) SetPipelines(pipelines []*pipeline.Pipeline) {
	for _, p := range pipelines {
		s.pipelines[p.Name] = p
	}
}

// Listen for calls to hook
func (s *HookServer) Listen() {
	http.Handle("/hook/", http.HandlerFunc(s.handleWebhook))
	if s.port != 0 {
		fmt.Printf("Listening on port %v\n", s.port)
		http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
		return
	}
}

// Function that handles weebooks by triggering the pipeline
// with the id set in the url
func (s *HookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {

	id, err := getPipelineId(r)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	// TODO : ADD AUTHENTICATION
	_, _, err = getBody(r.Body)
	defer r.Body.Close()

	//TODO add logging to MongoDB or json files

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	go s.BeginPipeline(id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received and verified"))
}

// Modified BeginPipeline to track active pipelines
func (s *HookServer) BeginPipeline(id string) {
    pipeline, ok := s.pipelines[id]
    if !ok {
        fmt.Printf("Wrong id received %s", id)
        return
    }
    
    // Create a new context for this pipeline execution
    ctx, cancelPipeline := context.WithCancel(context.Background())
    
    // Generate a unique execution ID
    executionID := fmt.Sprintf("%s:%s", pipeline.Name, pipeline.GetId())
    
    // Store the cancel function
    s.activePipelines.Store(executionID, cancelPipeline)
    
    // Create channel for cleanup coordination
    done := make(chan struct{})
    
    go func() {
        defer close(done)
        defer s.activePipelines.Delete(executionID)
        defer cancelPipeline()
        
        clone := *pipeline
        err := clone.ExecutePipeline(ctx)
        if err != nil {
            if err == context.Canceled {
                fmt.Printf("Pipeline '%s' was cancelled\n", pipeline.Name)
            } else {
                fmt.Printf("Pipeline '%s' failed with error: %v\n", pipeline.Name, err)
            }
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

// Returns the name of the pipeline to start
func getPipelineId(r *http.Request) (string, error) {

	segments := strings.Split(r.URL.Path, "/")

	// Expected path: "/users/{id}"
	if len(segments) < 3 || segments[1] != "hook" {
		return "", errors.New("Invalid url")
	}

	id := segments[2]
	return id, nil
}

// Returns the body of the http request as a struct and an array
// of bytes
func getBody(rBody io.ReadCloser) (WebhookPayload, []byte, error) {
	body, err := io.ReadAll(rBody)
	if err != nil {
		return WebhookPayload{}, nil, err
	}

	var payload WebhookPayload
	err = json.Unmarshal(body, &payload)
	return payload, body, err

}
