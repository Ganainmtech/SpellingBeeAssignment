package dictionary

import "strings"

// WithCache wraps a Repository and adds an in-memory cache.
func WithCache(repository Repository) Repository {
	return &cachedRepository{
		innerRepository: repository,
		cacheMemory:     map[string]bool{},
	}
}

type cachedRepository struct {
	innerRepository Repository
	cacheMemory     map[string]bool // lowercased word -> true/false
}

func (cachedRepo *cachedRepository) Contains(word string) bool {
	lowercaseWord := strings.ToLower(word)

	// Check cache first
	if cachedValue, exists := cachedRepo.cacheMemory[lowercaseWord]; exists {
		return cachedValue
	}

	// Not cached, ask the real repo and remember the result
	foundInRepository := cachedRepo.innerRepository.Contains(lowercaseWord)
	cachedRepo.cacheMemory[lowercaseWord] = foundInRepository
	return foundInRepository
}
