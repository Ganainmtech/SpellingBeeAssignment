package game

import (
	"math/rand"
	"strings"
	"unicode"
)

// LettersFromPangram takes a word and
// collects 7 letters then picks one as the center
func LettersFromPangram(p string) (letters []string, center string) {
	set := make(map[rune]bool)
	var uniq []rune

	// collect up to 7 unique letters
	for _, r := range strings.ToLower(p) {
		if unicode.IsLetter(r) && !set[r] {
			set[r] = true
			uniq = append(uniq, r)
			if len(uniq) == 7 {
				break
			}
		}
	}

	// if less than 7 unique found add more from the word
	for _, r := range strings.ToLower(p) {
		if len(uniq) == 7 {
			break
		}
		if unicode.IsLetter(r) {
			uniq = append(uniq, r)
		}
	}

	// convert runes to strings
	for _, r := range uniq[:7] {
		letters = append(letters, string(r))
	}

	// pick random center letter
	center = letters[rand.Intn(len(letters))]

	return letters, center
}
