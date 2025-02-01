package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Cyber-cicco/jerminal/pipeline"
)

// Test variable. Should be replaced with config
const TEST_ENV_VAR = "GITHUB_WEBHOOK_SECRET"

// Server receives the webhook call and executes pipelines
type Server struct {
	pipelines map[string]*pipeline.Pipeline // map of names to a pipeline
	port      uint16                       // port to listen to
}

// Creates a new server to Listen for incoming webhooks
func New(port uint16) *Server {
    return  &Server{
    	pipelines: map[string]*pipeline.Pipeline{},
    	port:      port,
    }
}

// Puts the pipelines in the server
func (s *Server) SetPipelines(pipelines []*pipeline.Pipeline) {
    for _, p := range pipelines {
        s.pipelines[p.Name] = p
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

    //TODO add logging to MongoDB or json files

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

    go s.BeginPipeline(id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received and verified"))
}

func (s *Server) BeginPipeline(id string) {
    pipeline, ok := s.pipelines[id] // pipeline is a pointer to a map of predefined pipelines
    if !ok {
        fmt.Printf("Wrong id received %s", id)
        return
    }
    clone := *pipeline
    err := clone.ExecutePipeline()
    fmt.Printf("err: %v\n", err)
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
