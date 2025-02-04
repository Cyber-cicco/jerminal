package pipeline

import "sync"

type Store struct {
    sync.Mutex
    activePipelines map[string]*Pipeline
    globalPipelines map[string]*Pipeline
}





