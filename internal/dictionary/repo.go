package dictionary

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Repository allows checking if a word exists
type Repository interface {
	Contains(word string) bool
}

// fileRepo stores words in memory for fast lookup
type fileRepo struct {
	words map[string]struct{} // map of lowercase words
}

// NewFileRepo loads words from a JSON object like {"word1":1, "word2":1}
func NewFileRepo(filePath string) (Repository, error) {
	// read file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// parse JSON object
	var rawDict map[string]interface{}
	if err := json.Unmarshal(fileBytes, &rawDict); err != nil {
		return nil, errors.New("expected JSON object of words")
	}

	// store keys in memory for fast lookup
	wordMap := make(map[string]struct{}, len(rawDict))
	for word := range rawDict {
		cleanedWord := strings.ToLower(strings.TrimSpace(word))
		wordMap[cleanedWord] = struct{}{}
	}

	return &fileRepo{words: wordMap}, nil
}

// Contains returns true if the word exists in the dictionary
func (repo *fileRepo) Contains(word string) bool {
	_, exists := repo.words[strings.ToLower(word)]
	return exists
}
