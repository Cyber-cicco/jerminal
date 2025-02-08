package dst

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
)

const (
	BOOL_TYPE = iota
	STRING_TYPE
	NUMBER_TYPE
)

func randBytes(src *rand.Rand, t *testing.T) []byte {
	newSeed := rand.New(rand.NewSource(src.Int63()))
	maxLen := src.Intn(10000)
	bytes := make([]byte, maxLen)
	for i := 0; i < maxLen; i++ {
		bytes = append(bytes, byte(newSeed.Intn(256)))
	}
	return bytes
}

// Helper function for reading the complete response
func readFullResponse(conn net.Conn) ([]byte, error) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(rpc.SplitFunc)
	scanner.Scan()
	msg := scanner.Bytes()
	_, content, err := rpc.DecodeMessage[rpc.JRPCResponse](msg)
	if err != nil {
		fmt.Printf("req: %v\n", string(content))
		return nil, err
	}
	return content, nil
}

func randomMap(src *rand.Rand, t *testing.T) map[string]any {
	maxLen := src.Intn(50)
	randomMap := make(map[string]any, maxLen)
	for i := 0; i < maxLen; i++ {
		paramType := randType(src)
		switch paramType {
		case STRING_TYPE:
			randomMap[randString(src)] = randString(src)
		case BOOL_TYPE:
			randomMap[randString(src)] = randBool(src)
		case NUMBER_TYPE:
			randomMap[randString(src)] = src.Int63()
		}
	}
	return randomMap
}

func randType(src *rand.Rand) uint8 {
	return uint8(src.Intn(4))
}

func randBool(src *rand.Rand) bool {
	return src.Intn(2) == 1
}

func randString(src *rand.Rand) string {
	strLen := src.Intn(50)
	bytes := make([]byte, strLen)
	src.Read(bytes)
	return string(bytes)
}

// DST is a package for deterministic simulation
// testing
// It uses the pipeline struct to test itself, since it is an
// appropriate framework to do so.
func TestSocketsProcesses(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("Seed for this test : %v", seed)

	src := rand.New(rand.NewSource(seed))
	p, err := SetPipeline("dst_scokets",
		Agent("dst"),
		// Test to see if random payloads produce unexpected results
		Stages("chaos_monkey",
			// See if I can crash the server with random bytes
			Stage("random_bullshit_go",
				Exec(func(p *Pipeline, ctx context.Context) error {
					for i := 0; i < 1000; i++ {
						conn, err := net.Dial("unix", "/tmp/pipeline-control.sock")
						if err != nil {
							t.Fatalf("unexpected error %v", err)
							panic(err)
						}
						defer conn.Close()
						bytes := randBytes(src, t)
						_, err = conn.Write(bytes)

						if err != nil {
							t.Fatalf("unexpected error %v", err)
							return err
						}

					}
					return nil
				}),
			),
			// See if it can handle every type of payloads, even invalid json
			Stage("random_payload_go",
				Exec(func(p *Pipeline, ctx context.Context) error {
					for i := 0; i < 1000; i++ {
						conn, err := net.Dial("unix", "/tmp/pipeline-control.sock")
						if err != nil {
							panic(err)
						}
						defer conn.Close()
						req := rpc.JRPCRequest{
							JsonRpcVersion: "2.0",
							Id:             0,
							Method:         "pipeline-cancelation",
						}
						params := rpc.CustomJRPCRequest{
							JRPCRequest: req,
							Params:      randomMap(src, t),
						}
						json, err := json.Marshal(params)
						if err != nil {
							t.Fatalf("unexpected error %v", err)
							return err
						}
						bytes := rpc.JRPCRes(json)
						fmt.Printf("payload :\n%v\n", string(bytes))
						_, err = conn.Write(bytes)
						if err != nil {
							t.Fatalf("unexpected error %v", err)
							return err
						}

						// Read and handle response
						response, err := readFullResponse(conn)
						if err != nil {
							t.Fatalf("unexpected error %v", err)
							return fmt.Errorf("failed to read response: %v", err)
						}
						fmt.Printf("response: %v\n", string(response))
					}

					return nil
				}),
			),
		),
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = p.ExecutePipeline(context.Background())
	if err != nil {
		t.Fatalf("Should not have gotten error, got %v", err)
	}
	if p.Inerror {
		p.Diagnostic.Log()
		t.Fatalf("Pipeline got error")
	}
}
