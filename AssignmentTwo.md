# Spelling Bee (SOFT8023 – Assignment 2)

Name: Sara Jane Kenny  
Student Number: R00255434  
Class: SDH3-A  

This assignment extends the Spelling Bee game from Assignment 1 by adding a socket-based statistics system, RabbitMQ event broadcasting, concurrency support, and Dockerised server components. The original gRPC game server and client continue to operate as in Assignment 1, with additional logic integrated for statistics and pangram events.

## Running the system

Start RabbitMQ using Docker:
```bash
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=password -e RABBITMQ_DEFAULT_VHOST=/ rabbitmq:management
```

Build Docker images:
```bash
docker build -f Dockerfile.server -t bee-server .
docker build -f Dockerfile.socket -t stats-server .
```

Run the gRPC Bee server (Docker):
```bash
docker run --rm --name bee-server -p 50051:50051 bee-server
```

Run the stats socket server (Docker):
```bash
docker run --rm --name stats-server -p 9000:9000 stats-server
```

Run the clients locally:
```bash
go run ./cmd/client
go run ./socket/client
```
Multiple socket clients may be run at the same time.

## Features implemented

1. **Socket-based statistics server**  
   A new server in socket/server listens on port 9000. It reads from data/stats.json, which is updated by the gRPC server. The socket server supports the following commands: HELP, TOTALGAMES, VALIDWORDS, PANGRAMS, HIGHESTSCORE, STATS/ALL, QUIT/EXIT. It sends stat results back to each connected client over TCP.

2. **File-based statistics management**  
   internal/stats.Manager maintains statistics in data/stats.json, including: total games started, total valid words, total pangrams, highest score. These values are incrementally updated during gameplay in the gRPC server.

3. **Pangram event broadcasting with RabbitMQ**  
   When a pangram is detected in the gRPC server's SubmitWord method, a message is published to the RabbitMQ exchange bee.events using routing key bee.pangram.  
   Example message: `Pangram 'factory' found in game g-1 (new total score: 15)`  
   Publishing is handled through a small helper in internal/mq/publisher.go.

4. **Pangram event consumption in socket clients**  
   Each socket client (socket/client) starts a goroutine that: connects to RabbitMQ, creates a temporary auto-deleted queue, binds the queue to bee.pangram, receives live pangram event messages, and prints them immediately on screen. All connected clients receive every pangram event (broadcast behaviour).

5. **Concurrency support**  
   The socket server handles clients concurrently using: `go handleClient(conn)`. Each client receives its own dedicated goroutine. The server supports multiple active clients, responding to stat queries and receiving RabbitMQ events independently.

6. **Dockerisation**  
   Two Dockerfiles were created:
   - Dockerfile.server – builds and runs the gRPC Bee server on port 50051
   - Dockerfile.socket – builds and runs the stats socket server on port 9000
   
   RabbitMQ is run separately using the rabbitmq:management image. Clients are executed locally and communicate with the Dockerised servers.

## Project structure (Assignment 2 additions)

- `socket/server/main.go` – stats TCP server
- `socket/client/main.go` – interactive TCP client with RabbitMQ event listener
- `internal/stats/stats.go` – statistics storage and update logic
- `internal/mq/publisher.go` – RabbitMQ publisher for pangram events
- `Dockerfile.server` – container build for gRPC server
- `Dockerfile.socket` – container build for stats server
- `data/stats.json` – stored statistics file

## Resources

Lab sheets and lecture material for sockets, concurrency, and RabbitMQ