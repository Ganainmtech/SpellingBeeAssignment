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
	"assignmentOne/internal/stats"
	"assignmentOne/internal/mq"
	"fmt"
	"os"

	"google.golang.org/grpc"
)

// main game server
type server struct {
	beepb.UnimplementedBeeServiceServer
	reg      *game.Registry        // game state manager
	dic      dictionary.Repository // dictionary with cache
	pangrams []string 
	stats *stats.Manager // handles reading/writing stats.json
	pub *mq.Publisher // RMQ publisher for pangram events
}

// make new server with given dictionary and pangrams
func newServer(dic dictionary.Repository, pangrams []string, statsMgr *stats.Manager, pub *mq.Publisher) *server {
	return &server{
		reg:      game.GetRegistry(),
		dic:      dic,
		pangrams: pangrams,
		stats: statsMgr,
		pub: pub,
	}
}

func (beeServer *server) StartGame(ctx context.Context, req *beepb.StartGameRequest) (*beepb.StartGameResponse, error) {
	var chosenPangram string
	pangramSeed := strings.TrimSpace(strings.ToLower(req.GetPangramSeed()))

	// pick pangram: seeded or random
	if pangramSeed != "" {
		log.Printf("using seed: %s", pangramSeed)
		chosenPangram = pangramSeed
	} else if len(beeServer.pangrams) > 0 {
		rand.Seed(time.Now().UnixNano())
		chosenPangram = beeServer.pangrams[rand.Intn(len(beeServer.pangrams))]
		log.Printf("random pangram: %s", chosenPangram)
	} else {
		log.Printf("no pangrams found")
	}

	// make 7 letters and a center
	letters, center := game.LettersFromPangram(chosenPangram)
	shuffleStrings(letters)

	// store game in registry
	gameId := beeServer.reg.Add(&game.State{
		Letters: letters,
		Center:  center,
		Score:   0,
		Found:   make(map[string]bool),
	})

	if beeServer.stats != nil {
		if err := beeServer.stats.RecordGameStarted(); err != nil {
			log.Printf("stats: failed to record game start: %v", err)
		}
	}

	return &beepb.StartGameResponse{
		GameId:       gameId,
		Letters:      letters,
		CenterLetter: center,
		CurrentScore: 0,
	}, nil
}

func (beeServer *server) SubmitWord(ctx context.Context, req *beepb.SubmitWordRequest) (*beepb.SubmitWordResponse, error) {
	gameState, ok := beeServer.reg.Get(req.GetGameId())
	if !ok {
		return &beepb.SubmitWordResponse{Valid: false, Reason: "game not found", TotalScore: 0}, nil
	}

	word := norm(req.GetWord())

	invalid := func(reason string) *beepb.SubmitWordResponse {
		return &beepb.SubmitWordResponse{Valid: false, Reason: reason, TotalScore: int32(gameState.Score)}
	}

	// quick checks
	if utf8.RuneCountInString(word) < 4 {
		return invalid("too short"), nil
	}
	if !containsCenter(word, gameState.Center) {
		return invalid("missing center"), nil
	}
	if !onlyFromSet(word, gameState.Letters) {
		return invalid("bad letters"), nil
	}
	if !beeServer.dic.Contains(word) {
		return invalid("not in dictionary"), nil
	}
	if gameState.Found[word] {
		return invalid("already found"), nil
	}

	// valid word gets score
	pangram := usesAllSeven(word, gameState.Letters)
	wordScore := scoreWord(word, pangram)

	gameState.Found[word] = true
	gameState.Score += wordScore

	// update stats for a valid word
	if beeServer.stats != nil {
		if err := beeServer.stats.RecordWordSubmitted(true, pangram, wordScore, gameState.Score); err != nil {
			log.Printf("stats: failed to record word: %v", err)
		}
	}

	// if word is pangram, broadcast an event via RMQ
	if pangram && beeServer.pub != nil {
		msg := fmt.Sprintf(
			"Pangram '%s' found in game %s (new total score: %d)",
			word, req.GetGameId(), gameState.Score,
		)
		beeServer.pub.Publish("bee.pangram", msg)
	}

	return &beepb.SubmitWordResponse{
		Valid:      true,
		WordScore:  int32(wordScore),
		Pangram:    pangram,
		TotalScore: int32(gameState.Score),
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
	pangramList, err := dictionary.LoadPangrams("data/pangrams.txt")
	if err != nil {
		log.Printf("warning: load pangrams: %v", err)
		pangramList = nil
	}
	log.Printf("loaded %d pangrams", len(pangramList))

	// create stats manager
	statsMgr, err := stats.NewManager("data/stats.json")
	if err != nil {
		log.Fatalf("init stats manager: %v", err)
	}

	// connect RMQ for pangram events
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://admin:password@localhost:5672/"
	}

	pub, err := mq.Connect(amqpURL)
	if err != nil {
		log.Printf("warning: could not connect to RabbitMQ (%v); pangram events disable", err)
		pub = nil
	} else {
		log.Printf("connected to RabbitMQ at %s", amqpURL)
	}
	if pub != nil {
		defer pub.Close()
	}

	// start gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	beepb.RegisterBeeServiceServer(grpcServer, newServer(dic, pangramList, statsMgr, pub))

	log.Printf("Bee server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// helper functions
// trim + lowercase as its a user input
func norm(inputWord string) string { return strings.ToLower(strings.TrimSpace(inputWord)) }

// check if word has center letter
func containsCenter(word, center string) bool {
	return strings.ContainsRune(word, []rune(center)[0])
}

// true if all letters in word come from allowed set
func onlyFromSet(word string, letters []string) bool {
	set := map[rune]bool{}
	for _, letter := range letters {
		set[[]rune(letter)[0]] = true
	}
	for _, runeValue := range word {
		if !set[runeValue] {
			return false
		}
	}
	return true
}

// true if word uses all 7 letters
func usesAllSeven(word string, letters []string) bool {
	need := map[rune]bool{}
	for _, letter := range letters {
		need[[]rune(letter)[0]] = true
	}
	for _, runeValue := range word {
		delete(need, runeValue)
		if len(need) == 0 {
			return true
		}
	}
	return false
}

// calc word score, +7 if pangram
func scoreWord(word string, pangram bool) int {
	wordLength := utf8.RuneCountInString(word)
	if wordLength == 4 {
		if pangram {
			return 1 + 7
		}
		return 1
	}
	score := wordLength
	if pangram {
		score += 7
	}
	return score
}

// shuffle order of letters
func shuffleStrings(letters []string) {
	rand.Seed(time.Now().UnixNano())
	for i := len(letters) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		letters[i], letters[j] = letters[j], letters[i]
	}
}
