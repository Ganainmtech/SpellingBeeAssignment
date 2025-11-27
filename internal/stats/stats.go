package stats

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Stats struct {
	TotalGames int `json:"total_games"` // how many games have started
	TotalValidWords int `json:"total_valid_words"`
	TotalPangrams int `json:"total_pangrams"` // how many pangrams were found
	HighestScore int `json:"highest_score"` // highest total score across games
}

// wraps Stats and handles concurrent access
type Manager struct {
	mu sync.Mutex // protects data
	path string
	data Stats
}

// loads exisiting stats or starts with empty/creates the file
func NewManager(path string) (*Manager, error) {
	m := &Manager{
		path: path,
		data: Stats{},
	}

	// try to load existing stats file (if it exists)
	if err := m.load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		// if no existing file, write an initial empty stats file
		if err := m.save(); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// load reads the stats JSON file from disk into m.data
// Called only during initialisation so we don't lock here
func (m *Manager) load() error {
	raw, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	var s Stats
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	m.data = s
	return nil
}

// save writes the current stats to disk as pretty-printed JSON
func (m *Manager) save() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, raw, 0o644)
}

// snapshot returns a copy of the current stats.
func (m *Manager) Snapshot() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data
}

// recordGameStarted increments the TotalGames counter and saves
func (m *Manager) RecordGameStarted() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data.TotalGames++
	return m.save()
}

// recordWordSubmitted is called every time a word is submitted
func (m *Manager) RecordWordSubmitted(isValid bool, isPangram bool, wordScore int, newTotalScore int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isValid { 
		m.data.TotalValidWords++ // updated only if isValid == true
		if isPangram {
			m.data.TotalPangrams++ // updated if isValid && isPangram
		}
		if newTotalScore > m.data.HighestScore { // updated if newTotalScore beats previous record
			m.data.HighestScore = newTotalScore
		}
	}

	return m.save()

}
