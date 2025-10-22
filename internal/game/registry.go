package game

import (
	"fmt"
	"sync"
)

// State is the per-game state the server mutates
type State struct {
	Letters []string
	Center  string
	Score   int
	Found   map[string]bool // prevent duplicate words from user
}

// Registry is a process-wide store of games (Singleton)
type Registry struct {
	mu    sync.RWMutex
	games map[string]*State
	next  int
}

var reg = &Registry{games: map[string]*State{}}

// GetRegistry returns the single, shared registry
func GetRegistry() *Registry { return reg }

// Add inserts a new game and returns its ID ("g-1" etc)
func (r *Registry) Add(g *State) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next++
	id := fmt.Sprintf("g-%d", r.next)
	r.games[id] = g
	return id
}

// Get looks up a game by ID
func (r *Registry) Get(id string) (*State, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.games[id]
	return g, ok
}
