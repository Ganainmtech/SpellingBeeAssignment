package dictionary

import (
	"bufio"
	"os"
)

// LoadPangrams reads one pangram per line from the file.
func LoadPangrams(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		word := sc.Text()
		if word != "" {
			out = append(out, word)
		}
	}
	return out, sc.Err()
}
