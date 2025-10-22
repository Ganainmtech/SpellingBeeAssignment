# Spelling Bee (SOFT8023 – Assignment 1)

Name: Sara Jane Kenny  
Student Number: R00255434  
Class: SDH3-A  

Single-player Spelling Bee game built in Go with gRPC.  
Client sends guesses to the server, which validates, scores, and returns an updated total.

## Running the game
- Terminal 1, Generate Go gRPC code from the proto: `protoc --go_out=. --go-grpc_out=. api/spelling/v1/bee.proto`
- Terminal 1: `go run ./cmd/server`
- Terminal 2: `go run ./cmd/client`

## Design Patterns
1. **Adapter**  
   - Where: internal/dictionary/repo.go  
   - What: Hides external dictionary format (JSON) and exposes a simple `Contains(word)` interface.  
   - Why: Server code doesn’t need to know file format, swapping the data source later is easy.  

2. **Factory**  
   - Where: internal/game/letters.go / server main  
   - What: Converts a pangram into exactly 7 letters and chooses a center letter.  
   - Why: Centralises letter/center generation, makes tests reproducible and logic consistent.  

3. **Decorator**  
   - Where: internal/dictionary/cache.go  
   - What: Wraps repository with in-memory caching.  
   - Why:  Dictionary lookups happen frequently, caching improves performance without changing adapter code.  

4. **Singleton**  
   - Where: internal/game/registry.go  
   - What: Single shared registry stores all game state.  
   - Why: Keeps one source of truth for multiple games, simplifies state management, preparation for assignment 2. 

## Project structure
- `api/spelling/v1/bee.proto` – gRPC contract  
- `cmd/server/main.go` – server  
- `cmd/client/main.go` – CLI client  
- `internal/dictionary/` – dictionary adapter, cache, pangram loader  
- `internal/game/registry.go` – singleton game store  
- `data/words_dictionary.json` – dictionary  
- `data/pangrams.txt` – pangrams list

## Resources
- Lab sheets & lecture slides for implementations for gRPC & design principles
- GeeksforGeeks website & genAI to understand design patterns and gRPC better
