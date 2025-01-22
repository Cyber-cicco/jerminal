package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Cyber-cicco/jerminal/pipeline"
)

// Server receives the webhook call and executes pipelines
type Server struct {
	port      uint16                       // port to listen to
	pipelines map[string]*pipeline.Pipeline // map of names to a pipeline
}

func (s *Server) Listen() {
    http.Handle("/hook/", http.HandlerFunc(listenToHooks))
	if s.port != 0 {
        http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
        return
	}
}

func listenToHooks(w http.ResponseWriter, r *http.Request) {
	// Split the URL path into segments
	segments := strings.Split(r.URL.Path, "/")

	// Expected path: "/users/{id}"
	if len(segments) < 3 || segments[1] != "hook" {
		http.NotFound(w, r)
		return
	}

	id := segments[2]

    //TODO : add logic
    fmt.Printf("id: %v\n", id)
}

func (s *Server) BeginPipeline(name string) {

}
