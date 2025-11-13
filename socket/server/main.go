package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"assignmentOne/internal/stats"
)

const (
	tcpAddr = ":9000" // addr for socket server
	statsPath = "data/stats.json"
)

// initialied once in main & then reused by goroutines
var statsMgr *stats.Manager

func main() {
	// initialise the stats manager
	var err error
	statsMgr, err = stats.NewManager(statsPath)
	if err != nil {
		log.Fatalf("init stats manager: %v", err)
	}
	log.Printf("stats socket server using %s", statsPath)

	// start tcp server
	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("Stats socket server listening on %s", tcpAddr)

	// each connection in its own goroutine, many clients can talk to server concurrently
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleClient(conn)
	}
}

// deal with single tcp client connection
func handleClient(conn net.Conn) {
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// message when client connects
	_, _ = w.WriteString("Welcome to the Spelling Bee stats server!\n")
	_ = w.Flush()

	for {
		// read one line, until client hit enter or disconnects
		line, err := r.ReadString('\n')
		if err != nil {
			// client disconnected or network error, end the goroutine
			return
		}

		cmd := strings.TrimSpace(line)
		if cmd == "" {
			// ignore empty lines
			continue
		}

		// allow case senstive
		cmdUpper := strings.ToUpper(cmd)

		switch cmdUpper {
		case "HELP":
			writeHelp(w)
		case "TOTALGAMES":
			writeTotalGames(w)
		case "VALIDWORDS":
			writeTotalValidWords(w)
		case "PANGRAMS":
			writeTotalPangrams(w)
		case "HIGHESTSCORE":
			writeHighestScore(w)
		case "ALL":
			writeAllStats(w)
		case "QUIT":
			_, _ = w.WriteString("Goodbye!\n")
			_ = w.Flush()
			return
		default:
			_, _ = w.WriteString("Unrecognised command. Try: HELP\n")
			_ = w.Flush()
		}
	}
}

// writeHelp prints a short list of supported commands
func writeHelp(w *bufio.Writer) {
	_, _ = w.WriteString("Available commands:\n")
	_, _ = w.WriteString("HELP: Show this help message\n")
	_, _ = w.WriteString("TOTALGAMES: Total number of games started\n")
	_, _ = w.WriteString("VALIDWORDS: Total number of valid words found\n")
	_, _ = w.WriteString("PANGRAMS: Total number of pangrams found\n")
	_, _ = w.WriteString("HIGHESTSCORE: Highest score achieved in any game\n")
	_, _ = w.WriteString("ALL: Print all stats in one line\n")
	_, _ = w.WriteString("QUIT: Close the connection\n")
	_ = w.Flush()
}

// Helper to get a snapshot once per command
func currentSnapshot() stats.Stats {
	return statsMgr.Snapshot()
}

func writeTotalGames(w *bufio.Writer) {
	s := currentSnapshot()
	_, _ = fmt.Fprintf(w, "Total games started: %d\n", s.TotalGames)
	_ = w.Flush()
}

func writeTotalValidWords(w *bufio.Writer) {
	s := currentSnapshot()
	_, _ = fmt.Fprintf(w, "Total valid words: %d\n", s.TotalValidWords)
	_ = w.Flush()
}

func writeTotalPangrams(w *bufio.Writer) {
	s := currentSnapshot()
	_, _ = fmt.Fprintf(w, "Total pangrams: %d\n", s.TotalPangrams)
	_ = w.Flush()
}

func writeHighestScore(w *bufio.Writer) {
	s := currentSnapshot()
	_, _ = fmt.Fprintf(w, "Highest score so far: %d\n", s.HighestScore)
	_ = w.Flush()
}

func writeAllStats(w *bufio.Writer) {
	s := currentSnapshot()
	_, _ = fmt.Fprintf(
		w,
		"Games: %d | Valid words: %d | Pangrams: %d | Highest score: %d\n",
		s.TotalGames, s.TotalValidWords, s.TotalPangrams, s.HighestScore,
	)
	_ = w.Flush()
}