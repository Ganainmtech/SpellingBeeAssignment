package game

import (
	"fmt"
	"sync"
)

// State is the per-game state the server mutates.
type State struct {
	Letters []string
	Center  string
	Score   int
	Found   map[string]bool // prevent duplicate words from user
}

// Registry is a process-wide store of games (Singleton).
type Registry struct {
	mu    sync.RWMutex
	games map[string]*State
	next  int
}

var sharedRegistry = &Registry{games: map[string]*State{}}

// GetRegistry returns the single, shared registry.
func GetRegistry() *Registry {
	return sharedRegistry
}

// Add inserts a new game and returns its ID ("g-1", etc.).
func (registry *Registry) Add(gameState *State) string {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.next++
	gameID := fmt.Sprintf("g-%d", registry.next)
	registry.games[gameID] = gameState
	return gameID
}

// Get looks up a game by ID.
func (registry *Registry) Get(gameID string) (*State, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	gameState, exists := registry.games[gameID]
	return gameState, exists
}
