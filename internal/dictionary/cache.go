package dictionary

import "strings"

// WithCache wraps a Repository and adds an in-memory cache.
func WithCache(inner Repository) Repository {
	return &cache{inner: inner, mem: map[string]bool{}}
}

type cache struct {
	inner Repository
	mem   map[string]bool // lowercased word -> true/false
}

func (c *cache) Contains(word string) bool {
	w := strings.ToLower(word)

	// Check cache first
	if v, ok := c.mem[w]; ok {
		return v
	}

	// Not cached: ask the real repository and remember the result
	v := c.inner.Contains(w)
	c.mem[w] = v
	return v
}
