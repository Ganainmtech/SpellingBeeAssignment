package main

import (
	beepb "assignmentOne/api/spelling/v1"
	"assignmentOne/internal/dictionary"
	"assignmentOne/internal/game"
	"context"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc"
)

// main game server
type server struct {
	beepb.UnimplementedBeeServiceServer
	reg      *game.Registry        // game state manager
	dic      dictionary.Repository // dictionary with cache
	pangrams []string              // list of pangrams for new games
}

// make new server with given dictionary and pangrams
func newServer(dic dictionary.Repository, pangrams []string) *server {
	return &server{
		reg:      game.GetRegistry(),
		dic:      dic,
		pangrams: pangrams,
	}
}

func (s *server) StartGame(ctx context.Context, req *beepb.StartGameRequest) (*beepb.StartGameResponse, error) {
	var chosen string
	seed := strings.TrimSpace(strings.ToLower(req.GetPangramSeed()))

	// pick pangram: seeded or random
	if seed != "" {
		log.Printf("using seed: %s", seed)
		chosen = seed
	} else if len(s.pangrams) > 0 {
		rand.Seed(time.Now().UnixNano())
		chosen = s.pangrams[rand.Intn(len(s.pangrams))]
		log.Printf("random pangram: %s", chosen)
	} else {
		log.Printf("no pangrams found")
	}

	// make 7 letters and a center
	letters, center := game.LettersFromPangram(chosen)
	shuffleStrings(letters)

	// store game in registry
	id := s.reg.Add(&game.State{
		Letters: letters,
		Center:  center,
		Score:   0,
		Found:   make(map[string]bool),
	})

	return &beepb.StartGameResponse{
		GameId:       id,
		Letters:      letters,
		CenterLetter: center,
		CurrentScore: 0,
	}, nil
}

func (s *server) SubmitWord(ctx context.Context, req *beepb.SubmitWordRequest) (*beepb.SubmitWordResponse, error) {
	g, ok := s.reg.Get(req.GetGameId())
	if !ok {
		return &beepb.SubmitWordResponse{Valid: false, Reason: "game not found", TotalScore: 0}, nil
	}

	word := norm(req.GetWord())

	invalid := func(reason string) *beepb.SubmitWordResponse {
		return &beepb.SubmitWordResponse{Valid: false, Reason: reason, TotalScore: int32(g.Score)}
	}

	// quick checks
	if utf8.RuneCountInString(word) < 4 {
		return invalid("too short"), nil
	}
	if !containsCenter(word, g.Center) {
		return invalid("missing center"), nil
	}
	if !onlyFromSet(word, g.Letters) {
		return invalid("bad letters"), nil
	}
	if !s.dic.Contains(word) {
		return invalid("not in dictionary"), nil
	}
	if g.Found[word] {
		return invalid("already found"), nil
	}

	// valid word gets score 
	pangram := usesAllSeven(word, g.Letters)
	ws := scoreWord(word, pangram)

	g.Found[word] = true
	g.Score += ws

	return &beepb.SubmitWordResponse{
		Valid:      true,
		WordScore:  int32(ws),
		Pangram:    pangram,
		TotalScore: int32(g.Score),
	}, nil
}

func main() {
	// load dictionary
	dic, err := dictionary.NewFileRepo("data/words_dictionary.json")
	if err != nil {
		log.Fatalf("load dictionary: %v", err)
	}
	dic = dictionary.WithCache(dic)

	// load pangrams
	pangs, err := dictionary.LoadPangrams("data/pangrams.txt")
	if err != nil {
		log.Printf("warning: load pangrams: %v", err)
		pangs = nil
	}
	log.Printf("loaded %d pangrams", len(pangs))

	// start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	beepb.RegisterBeeServiceServer(s, newServer(dic, pangs))

	log.Printf("Bee server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// helper functions
// clean word: trim + lowercase
func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// check if word has center letter
func containsCenter(word, center string) bool {
	return strings.ContainsRune(word, []rune(center)[0])
}

// true if all letters in word come from allowed set
func onlyFromSet(word string, letters []string) bool {
	set := map[rune]bool{}
	for _, s := range letters {
		set[[]rune(s)[0]] = true
	}
	for _, r := range word {
		if !set[r] {
			return false
		}
	}
	return true
}

// true if word uses all 7 letters
func usesAllSeven(word string, letters []string) bool {
	need := map[rune]bool{}
	for _, s := range letters {
		need[[]rune(s)[0]] = true
	}
	for _, r := range word {
		delete(need, r)
		if len(need) == 0 {
			return true
		}
	}
	return false
}

// calc word score, +7 if pangram
func scoreWord(word string, pangram bool) int {
	n := utf8.RuneCountInString(word)
	if n == 4 {
		if pangram {
			return 1 + 7
		}
		return 1
	}
	score := n
	if pangram {
		score += 7
	}
	return score
}

// shuffle order of letters
func shuffleStrings(a []string) {
	rand.Seed(time.Now().UnixNano())
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
