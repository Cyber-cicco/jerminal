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

// Server receives the webhook call and executes pipelines
type Server struct {
	listener        net.Listener    // Unix socket listener
	port            uint16          // port to listen to
	activePipelines sync.Map        // map[string]context.CancelFunc
	store           *pipeline.Store //keeps track of the project pipelines activity
}

// New creates a new server to Listen for incoming webhooks
func New(port uint16) *Server {
	server := &Server{
		port:  port,
		store: pipeline.GetStore(),
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
	go server.listenSockets()

	return server
}

// # The structure of the message should be of JSON RPC, like for the LSPs
//
// This would give an interface for other local programs to interact with the process.
func (s *Server) listenSockets() {
	for {
		fmt.Printf("At beginning of listening for cancelation")
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
				fmt.Printf("Message scanned")
				msg := scanner.Bytes()
				req, content, err := rpc.DecodeMessage(msg)
				if err != nil {
					fmt.Printf("Error encountered : %s", err)
					continue
				}
				res, err := s.handleMessage(req, content)
				if err != nil {
					fmt.Printf("Could not marshall struct: %v\n", err)
				}
				_, err = c.Write(res)
				if err != nil {
					fmt.Println("Could not write to unix socket")
				}
			}

		}(conn)
	}
}

// Puts the pipelines in the server
func (s *Server) SetPipelines(pipelines []*pipeline.Pipeline) {
    s.store.Lock()
    defer s.store.Unlock()
	for _, p := range pipelines {
		s.store.GlobalPipelines[p.Name] = p
	}
}

// Listen for calls to hook
func (s *Server) Listen() {
	http.Handle("/hook/", http.HandlerFunc(s.handleWebhook))
	if s.port != 0 {
		fmt.Printf("Listening on port %v\n", s.port)
		http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
		return
	}
}

// Function that handles weebooks by triggering the pipeline
// with the id set in the url
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {

	id, err := getPipelineId(r)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	// TODO : ADD AUTHENTICATION
	_, _, err = getBody(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	go s.BeginPipeline(id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received and verified"))
}

// Modified BeginPipeline to track active pipelines
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

	// Generate a unique execution ID
	executionID := fmt.Sprintf("%s", clone.GetId())

	// Store the cancel function
	s.activePipelines.Store(executionID, cancelPipeline)

	// Create channel for cleanup coordination
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer s.activePipelines.Delete(executionID)
		defer cancelPipeline()

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
