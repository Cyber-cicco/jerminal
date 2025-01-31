package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Cyber-cicco/jerminal/pipeline"
)

// Test variable
const TEST_ENV_VAR = "GITHUB_WEBHOOK_SECRET"

// Server receives the webhook call and executes pipelines
type Server struct {
	pipelines map[string]*pipeline.Pipeline // map of names to a pipeline
	port      uint16                       // port to listen to
}

func New(port uint16) *Server {
    return  &Server{
    	pipelines: map[string]*pipeline.Pipeline{},
    	port:      port,
    }
}

// Listen for calls to hook
func (s *Server) Listen() {
    http.Handle("/hook/", http.HandlerFunc(listenToHooks))
	if s.port != 0 {
        fmt.Printf("Listening on port %v\n", s.port)
        http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
        return
	}
}

func (s *Server) TestListen() {
    http.Handle("/hook", http.HandlerFunc(handleWebhook))
	if s.port != 0 {
        fmt.Printf("Listening on port %v\n", s.port)
        http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
        return
	}

}

// listenToHooks launches a pipeline identified by
// the id in the path
func listenToHooks(w http.ResponseWriter, r *http.Request) {
	// Split the URL path into segments
	segments := strings.Split(r.URL.Path, "/")

	// Expected path: "/users/{id}"
	if len(segments) < 3 || segments[1] != "hook" {
		http.NotFound(w, r)
		return
	}

	id := segments[2]

    fmt.Printf("id: %v\n", id)
}

// Test function written by llm to check that it works
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Retrieve the secret from the environment variable
	secret := os.Getenv(TEST_ENV_VAR)
	if secret == "" {
		http.Error(w, "Webhook secret not set", http.StatusInternalServerError)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
    fmt.Printf("body: %v\n", body)
	defer r.Body.Close()

	// Get the signature from the header
	signature := r.Header.Get("X-Hub-Signature")
	if signature == "" {
		http.Error(w, "No signature header", http.StatusBadRequest)
		return
	}

	// Verify the signature
	if !verifySignature(secret, signature, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Handle the payload
	fmt.Println("Webhook received and verified")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received and verified"))
}

func verifySignature(secret, signature string, body []byte) bool {
	// The signature is in the format "sha1=hash"
	const prefix = "sha1="

    if !strings.HasPrefix(signature, prefix) {
        return false
    }

	if len(signature) != len(prefix)+sha1.Size*2 || !hmac.Equal([]byte(signature[:len(prefix)]), []byte(prefix)) {
		return false
	}

	// Compute the HMAC
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)

	// Decode the signature
	actualMAC, err := hex.DecodeString(signature[len(prefix):])
	if err != nil {
		return false
	}

	// Compare the MACs
	return hmac.Equal(actualMAC, expectedMAC)
}

func (s *Server) BeginPipeline(name string) {

}
