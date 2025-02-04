package pipeline

import "sync"

type Store struct {
	sync.Mutex
	ActivePipelines map[string]*Pipeline
	GlobalPipelines map[string]*Pipeline
}

var store *Store

func GetStore() *Store {
	if store == nil {
		store = &Store{
			ActivePipelines: make(map[string]*Pipeline),
			GlobalPipelines: make(map[string]*Pipeline),
		}
	}
    return store
}
