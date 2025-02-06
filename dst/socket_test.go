package dst

import (
	"context"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"

	. "github.com/Cyber-cicco/jerminal/pipeline"
	"github.com/Cyber-cicco/jerminal/server/rpc"
)

func randBytes(src *rand.Rand, t *testing.T) []byte {
	newSeed := rand.New(rand.NewSource(src.Int63()))
	maxLen := src.Intn(10000)
	t.Logf("seed for this run : %v", newSeed)
	bytes := make([]byte, maxLen)
	for i := 0; i < maxLen; i++ {
		bytes = append(bytes, byte(newSeed.Intn(256)))
	}
	return bytes
}

// Helper function for reading the complete response
func readFullResponse(conn net.Conn) ([]byte, error) {
	var response []byte
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		response = append(response, buffer[:n]...)
	}

	return response, nil
}

// DST is a package for deterministic simulation
// testing
// It uses the pipeline struct to test itself, since it is an
// appropriate framework to do so.
func TestSocketsProcesses(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("Seed for this test : %v", seed)
	conn, err := net.Dial("unix", "/tmp/pipeline-control.sock")
	if err != nil {
		panic(err)
	}
	defer conn.Close() // Good practice to close the connection when done

	src := rand.New(rand.NewSource(seed))
	p, err := SetPipeline("dst_scokets",
		Agent("dst"),
        // Test to see if random payloads produce unexpected results
		Stages("chaos_monkey",
			Stage("random_bullshit_go",
				Exec(func(p *Pipeline, ctx context.Context) error {
					bytes := randBytes(src, t)
					_, err := conn.Write(bytes)
					if err != nil {
						return err
					}

					return nil
				}),
			).Retry(10000, time.Millisecond*1),
		).Parallel(),
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = p.ExecutePipeline(context.Background())
	if err != nil {
		t.Fatalf("Should not have gotten error, got %v", err)
	}
}
