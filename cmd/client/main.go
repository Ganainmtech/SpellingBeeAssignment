package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	beepb "assignmentOne/api/spelling/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// connect to server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// make gRPC client
	c := beepb.NewBeeServiceClient(conn)

	// start a new game
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start, err := c.StartGame(ctx, &beepb.StartGameRequest{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Spelling Bee!")
	printLetters(start.GetLetters(), start.GetCenterLetter())

	score := int(start.GetCurrentScore())
	in := bufio.NewScanner(os.Stdin)

	// read user words and send to server
	for {
		fmt.Print("Enter word > ")
		if !in.Scan() {
			break
		}
		w := strings.TrimSpace(strings.ToLower(in.Text()))
		if w == "" {
			continue
		}
		if w == "quit" || w == "exit" {
			break
		}

		// send one request and get one response
		resp, err := c.SubmitWord(context.Background(), &beepb.SubmitWordRequest{
			GameId: start.GetGameId(),
			Word:   w,
		})
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		// server handles scoring, client just shows results
		score = int(resp.GetTotalScore())
		if !resp.GetValid() {
			fmt.Printf("Sorry, %s  score: %d\n", resp.GetReason(), score)
		} else if resp.GetPangram() {
			fmt.Printf("Pangram! %d points  score: %d\n", resp.GetWordScore(), score)
		} else {
			fmt.Printf("Valid word +%d  score: %d\n", resp.GetWordScore(), score)
		}

		// show letters again after each guess
		printLetters(start.GetLetters(), start.GetCenterLetter())
	}
}

// printLetters shows letters like: A Y O [T] C U R
// the one in brackets is the required center
func printLetters(letters []string, center string) {
	c := strings.ToLower(center)
	parts := make([]string, 0, len(letters))
	for _, L := range letters {
		if strings.ToLower(L) == c {
			parts = append(parts, "["+strings.ToUpper(L)+"]")
		} else {
			parts = append(parts, strings.ToUpper(L))
		}
	}
	fmt.Println(strings.Join(parts, " "))
}
