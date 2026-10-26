# Spelling Bee

A client-server Spelling Bee game built with Go, developed as part of the SOFT8023 coursework. The project explores distributed systems, network communication, messaging, concurrency, and containerisation through two progressive assignments.

## Overview

The application implements a word puzzle inspired by the Spelling Bee format. Players submit words to a server, which validates guesses, calculates scores, and maintains game statistics.

The project evolves across two assignments:

- **Assignment 1:** A single-player game using gRPC, with dictionary validation, scoring, and shared game state.
- **Assignment 2:** Extends the system with a TCP socket-based statistics service, RabbitMQ event broadcasting, concurrent client handling, and Dockerised servers.

## Features

### Core game

- Client-server communication using gRPC and Protocol Buffers.
- Word validation against a dictionary.
- Word scoring and cumulative game scores.
- Seven-letter puzzles with a designated centre letter.
- Pangram detection.
- Shared game state and dictionary lookup caching.

### Statistics and messaging

- TCP socket server for querying game statistics.
- Tracks total games, valid words, pangrams, and highest score.
- Supports multiple simultaneous socket clients.
- RabbitMQ event broadcasting when a pangram is discovered.
- Live pangram notifications delivered to connected socket clients.
- JSON-based statistics storage.

### Containerisation

- Separate Docker images for the gRPC game server and TCP statistics server.
- RabbitMQ running as a separate service.
- Clients connect to the server components over their respective network ports.

## Tech Stack

- **Language:** Go
- **RPC:** gRPC and Protocol Buffers
- **Networking:** TCP sockets
- **Messaging:** RabbitMQ
- **Data storage:** JSON files
- **Concurrency:** Goroutines
- **Containerisation:** Docker

## Architecture

The application is organised into separate components for gameplay, statistics, and event messaging.

```text
                    +------------------+
                    |   Go Game Client |
                    +--------+---------+
                             |
                             | gRPC :50051
                             |
                    +--------v---------+
                    |   Game Server    |
                    |                  |
                    | Word Validation  |
                    | Scoring & State  |
                    | Statistics      |
                    +----+--------+----+
                         |        |
                 JSON stats       | Pangram events
                         |        |
                +--------v--+  +--v---------+
                | Statistics|  |  RabbitMQ  |
                |   JSON    |  |   Broker   |
                +--------+--+  +--+---------+
                         |        |
                         |        | Event notifications
                  TCP :9000        |
                         |        |
                +--------v--------v-+
                |  Socket Clients   |
                | Statistics & Live |
                | Pangram Events    |
                +-------------------+
```

## Project Structure

```text
SpellingBeeAssignment/
├── api/
│   └── spelling/v1/       # Protocol Buffers API definitions
├── assignmentOne/         # Assignment 1 API materials
├── cmd/
│   ├── server/            # gRPC game server
│   └── client/            # Command-line game client
├── data/
│   ├── words_dictionary.json
│   ├── pangrams.txt
│   └── stats.json
├── internal/
│   ├── dictionary/        # Dictionary access and caching
│   ├── game/              # Game logic and state management
│   ├── stats/             # Statistics management
│   └── mq/                # RabbitMQ event publishing
├── socket/
│   ├── server/            # TCP statistics server
│   └── client/            # TCP client and event listener
├── Dockerfile.server
├── Dockerfile.socket
├── AssignmentOne.md
├── AssignmentTwo.md
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

Install the following tools:

- Go 1.25 or a compatible version.
- Git.
- Protocol Buffers compiler (`protoc`).
- Go Protocol Buffers and gRPC code-generation plugins.
- Docker.
- RabbitMQ, for Assignment 2.

### Clone the repository

```bash
git clone https://github.com/Ganainmtech/SpellingBeeAssignment.git
cd SpellingBeeAssignment
```

### Install Go dependencies

```bash
go mod download
```

### Generate gRPC code

If the generated Protocol Buffers files need to be regenerated, install the Go plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Ensure the Go binary directory is on your `PATH`, then run:

```bash
protoc --go_out=. --go-grpc_out=. api/spelling/v1/bee.proto
```

### Run Assignment 1

Open two terminals from the repository root.

**Terminal 1 — Start the game server:**

```bash
go run ./cmd/server
```

**Terminal 2 — Start the game client:**

```bash
go run ./cmd/client
```

The client communicates with the server over gRPC to submit guesses and receive updated scores.

## Running Assignment 2

Assignment 2 adds the statistics socket server and RabbitMQ event messaging.

### 1. Start RabbitMQ

```bash
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=password \
  -e RABBITMQ_DEFAULT_VHOST=/ \
  rabbitmq:management
```

The RabbitMQ management interface is available at `http://localhost:15672` with the credentials configured above.

### 2. Build the server images

```bash
docker build -f Dockerfile.server -t bee-server .
docker build -f Dockerfile.socket -t stats-server .
```

### 3. Start the gRPC game server

```bash
docker run --rm --name bee-server -p 50051:50051 bee-server
```

### 4. Start the statistics server

In another terminal:

```bash
docker run --rm --name stats-server -p 9000:9000 stats-server
```

### 5. Start the clients

Run the game client:

```bash
go run ./cmd/client
```

Run the statistics client:

```bash
go run ./socket/client
```

Multiple socket clients can run concurrently, allowing statistics queries and pangram notifications to be handled by different clients.

## Statistics Commands

The socket statistics client supports the following commands:

| Command | Purpose |
| --- | --- |
| `HELP` | Display available commands |
| `TOTALGAMES` | View the number of games started |
| `VALIDWORDS` | View the total number of valid words |
| `PANGRAMS` | View the total number of pangrams |
| `HIGHESTSCORE` | View the highest recorded score |
| `STATS` / `ALL` | View statistics |
| `QUIT` / `EXIT` | Disconnect from the client |

## Design Patterns

The project applies several design patterns to keep the implementation modular and maintainable.

- **Adapter:** Provides a consistent interface for dictionary lookups, hiding the underlying JSON format.
- **Factory:** Centralises puzzle letter generation and centre-letter selection.
- **Decorator:** Adds in-memory caching around dictionary lookups without changing the underlying repository.
- **Singleton:** Maintains a shared game registry as a single source of truth for game state.

## Learning Outcomes

This project provided practical experience with:

- Designing client-server applications in Go.
- Defining APIs with Protocol Buffers and gRPC.
- Implementing TCP socket communication.
- Building asynchronous event-driven systems with RabbitMQ.
- Managing concurrent connections using goroutines.
- Separating application logic into modular components.
- Persisting application statistics using JSON.
- Containerising services with Docker.

## Coursework Documentation

The repository includes additional documentation for each stage of the project:

- [`AssignmentOne.md`](AssignmentOne.md) — Core game, gRPC implementation, and design patterns.
- [`AssignmentTwo.md`](AssignmentTwo.md) — Socket statistics, RabbitMQ messaging, concurrency, and Docker setup.

---

Developed as part of the SOFT8023 coursework.
