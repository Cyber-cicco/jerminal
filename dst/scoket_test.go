package dst

import (
	"context"
	"math/rand"
	"testing"
	"time"

	. "github.com/Cyber-cicco/jerminal/pipeline"
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
		Stages("chaos_monkey",
			Stage("random_bullshit_go",
				Exec(func(p *Pipeline, ctx context.Context) error {
                    randBytes(src, t)
					return nil
				}),
			).Retry(1000, time.Millisecond * 10),
		),
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	p.ExecutePipeline(context.Background())
}
