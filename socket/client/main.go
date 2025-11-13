package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Interactive client for the Spelling Bee stats socket server
func main() {
	// Default address
	addr := "localhost:9000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	// connect to the stats socket server.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	serverReader := bufio.NewReader(conn)
	serverWriter := bufio.NewWriter(conn)

	// continuously reads lines from the server & prints in console 
	// allows the server to send messages at any time without blocking the main thread
	go func() {
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				// Server closed or network error, exit
				fmt.Println("\nDisconnected from stats server.")
				os.Exit(0)
			}
			// Print the server message as-is (trim only trailing newline)
			fmt.Print(strings.TrimRight(line, "\n") + "\n")
			// Reprint prompt after async server message
			fmt.Print("> ")
		}
	}()

	// Main goroutine: simple REPL (Read-Eval-Print Loop)
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("Stats client connected to", addr)
	fmt.Println("Type HELP for commands, or QUIT to exit.")
	fmt.Print("> ")

	for input.Scan() {
		cmd := strings.TrimSpace(input.Text())
		if cmd == "" {
			fmt.Print("> ")
			continue
		}
		if strings.EqualFold(cmd, "quit") || strings.EqualFold(cmd, "exit") {
			// tell the server we're quitting.
			fmt.Fprintln(serverWriter, "QUIT")
			serverWriter.Flush()
			return
		}

		// Send the command line to the server (adds '\n' automatically)
		if _, err := fmt.Fprintln(serverWriter, cmd); err != nil {
			log.Fatalf("write: %v", err)
		}
		if err := serverWriter.Flush(); err != nil {
			log.Fatalf("flush: %v", err)
		}

		// Prompt again the background goroutine prints server responses
		fmt.Print("> ")
	}

	if err := input.Err(); err != nil {
		log.Fatalf("stdin: %v", err)
	}
}
