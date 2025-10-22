package dictionary

import (
	"bufio"
	"os"
)

// LoadPangrams reads one pangram per line from the file.
func LoadPangrams(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pangrams []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			pangrams = append(pangrams, line)
		}
	}
	return pangrams, scanner.Err()
}
