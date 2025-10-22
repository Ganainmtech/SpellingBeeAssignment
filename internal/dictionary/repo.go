package dictionary

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Repository lets us check if a word is in the dictionary
type Repository interface {
	Contains(word string) bool
}

// fileRepo keeps words in memory for fast lookup
type fileRepo struct {
	words map[string]struct{}
}

// NewFileRepo loads words from a JSON array like ["word1", "word2"]
func NewFileRepo(path string) (Repository, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rawArr []string
	if err := json.Unmarshal(b, &rawArr); err != nil {
		return nil, errors.New("expected JSON array of words")
	}

	m := make(map[string]struct{}, len(rawArr))
	for _, w := range rawArr {
		m[strings.ToLower(strings.TrimSpace(w))] = struct{}{}
	}

	return &fileRepo{words: m}, nil
}

// Contains checks if a word exists in the dictionary
func (r *fileRepo) Contains(word string) bool {
	_, ok := r.words[strings.ToLower(word)]
	return ok
}
