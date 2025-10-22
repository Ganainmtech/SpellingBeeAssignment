package game

import (
	"math/rand"
	"strings"
	"unicode"
)

// LettersFromPangram takes a word and
// collects 7 letters then picks one as the center.
func LettersFromPangram(pangramWord string) (letters []string, center string) {
	// Track which letters we hav already added (map) and store them in order (slice)
	uniqueRunes := make(map[rune]bool)
	var uniqueLetters []rune

	// collect up to 7 unique letters
	for _, letterRune := range strings.ToLower(pangramWord) {
		if unicode.IsLetter(letterRune) && !uniqueRunes[letterRune] {
			uniqueRunes[letterRune] = true
			uniqueLetters = append(uniqueLetters, letterRune)
			if len(uniqueLetters) == 7 {
				break
			}
		}
	}

	// if less than 7 unique found, add more from the word
	for _, letterRune := range strings.ToLower(pangramWord) {
		if len(uniqueLetters) == 7 {
			break
		}
		if unicode.IsLetter(letterRune) {
			uniqueLetters = append(uniqueLetters, letterRune)
		}
	}

	// convert runes to strings
	for _, letterRune := range uniqueLetters[:7] {
		letters = append(letters, string(letterRune))
	}

	// pick random center letter
	center = letters[rand.Intn(len(letters))]

	return letters, center
}
