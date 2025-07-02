package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Cyber-cicco/jerminal/config"
	"github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
)

// Server receives the webhook call and executes pipelines
type Server struct {
	listener        net.Listener                // Unix socket listener
	activePipelines sync.Map                    // map[string]context.CancelFunc
	store           *pipeline.Store             //keeps track of the project pipelines activity
	config          *config.GlobalStateProvider // constants of the process
}

// New creates a new server to Listen for incoming webhooks
func New() *Server {
	server := &Server{
		store: pipeline.GetStore(),
	}

	socketPath := "/tmp/pipeline-control.sock"
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Failed to create Unix socket: %v\n", err)
		return server
	}

	err = os.Chmod(socketPath, 0666)
	if err != nil {
		fmt.Printf("Failed to set socket permissions: %v\n", err)
	}

	conf, err := config.GetState()
	if err != nil {
		fmt.Printf("Server could not start because of error\n%v\n", err)
		os.Exit(1)
	}

	server.config = conf
	server.listener = listener

	// Start socket listener in goroutine
	go server.listenSockets()

	return server
}

// # The structure of the message should be of JSON RPC, like for the LSPs
//
// This would give an interface for other local programs to interact with the process.
func (s *Server) listenSockets() {
	for true {
		c, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		defer c.Close()

		scanner := bufio.NewScanner(c)
		scanner.Split(rpc.SplitFunc)
		for scanner.Scan() {
			fmt.Printf("Message scanned\n")
			msg := scanner.Bytes()
			req, content, err := rpc.DecodeMessage[rpc.JRPCRequest](msg)
			if err != nil {
				bytes := marshallError()
				_, err = c.Write(rpc.JRPCRes(bytes))
				if err != nil {
					fmt.Println("Could not write to unix socket")
				}
				continue
			}
			res := s.handleMessage(req, content)
			_, err = c.Write(rpc.JRPCRes(res))
			if err != nil {
				fmt.Println("Could not write to unix socket")
				continue
			}
		}
	}
}

// Puts the pipelines in the server
func (s *Server) SetPipelines(pipelines ...*pipeline.Pipeline) {
	s.store.Lock()
	defer s.store.Unlock()
	for _, p := range pipelines {
		s.store.GlobalPipelines[p.Name] = p
	}
}

// ListenGithubHooks for calls to hook
func (s *Server) ListenGithubHooks(port uint16) {
	http.Handle("/hook/github/", http.HandlerFunc(s.handleWebhook))
	if port != 0 {
		fmt.Printf("Listening on port %v\n", port)
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
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

	_, body, err := getBody(r.Body)
	defer r.Body.Close()

	verifyGithubSignature(s.config.GithubWebhookSecret, r.Header.Get("X-Hub-Signature"), body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	go s.BeginPipeline(id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received and verified"))
}

// Returns the name of the pipeline to start
func getPipelineId(r *http.Request) (string, error) {

	segments := strings.Split(r.URL.Path, "/")

	// Expected path: "/hook/github/{id}"
	if len(segments) < 4 || segments[1] != "hook" {
		return "", errors.New("Invalid url")
	}

	id := segments[3]
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
