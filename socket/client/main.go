package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

const defaultAMQPURL = "amqp://admin:password@localhost:5672/"

// return the AMQP URL from env if set
func getAMQPURL() string {
	if v := os.Getenv("AMQP_URL"); v != "" {
		return v
	}
	return defaultAMQPURL
}

//  RabbitMQ pangram consumer (runs in background goroutine)
func consumePangramEvents() {
	// connect to RabbitMQ
	conn, err := amqp.Dial(getAMQPURL())
	if err != nil {
		log.Printf("RabbitMQ dial error (client): %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("RabbitMQ channel error (client): %v", err)
		return
	}
	defer ch.Close()

	// declare the same exchange the server publishes to
	if err := ch.ExchangeDeclare(
		"bee.events", // name
		"topic",      // type
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // args
	); err != nil {
		log.Printf("RabbitMQ exchange declare (client): %v", err)
		return
	}

	// temporary auto-delete queue
	q, err := ch.QueueDeclare(
		"",    // let RabbitMQ pick a name
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("RabbitMQ queue declare (client): %v", err)
		return
	}

	// bind queue to pangram events
	if err := ch.QueueBind(
		q.Name,
		"bee.pangram", // routing key
		"bee.events",  // exchange
		false,
		nil,
	); err != nil {
		log.Printf("RabbitMQ queue bind (client): %v", err)
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,  // auto-ack
		true,  // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Printf("RabbitMQ consume (client): %v", err)
		return
	}

	// loop forever printing pangram events
	for d := range msgs {
		payload := strings.TrimSpace(string(d.Body))
		line := fmt.Sprintf("[PANGRAM EVENT] %s", payload)

		// print message then reprint prompt
		fmt.Print("\n" + line + "\n> ")
	}

	log.Println("Pangram event consumer ended")
}

//  Interactive stats socket client
func main() {
	// default address
	addr := "localhost:9000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	// connect to stats socket server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	serverReader := bufio.NewReader(conn)
	serverWriter := bufio.NewWriter(conn)

	// background goroutine to read messages from stats server
	go func() {
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("\nDisconnected from stats server.")
				os.Exit(0)
			}
			fmt.Print(strings.TrimRight(line, "\n") + "\n")
			fmt.Print("> ")
		}
	}()

	// start pangram event consumer
	go consumePangramEvents()

	// Main REPL
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
			fmt.Fprintln(serverWriter, "QUIT")
			serverWriter.Flush()
			return
		}

		_, err := fmt.Fprintln(serverWriter, cmd)
		if err != nil {
			log.Fatalf("write: %v", err)
		}

		if err := serverWriter.Flush(); err != nil {
			log.Fatalf("flush: %v", err)
		}

		fmt.Print("> ")
	}

	if err := input.Err(); err != nil {
		log.Fatalf("stdin: %v", err)
	}
}
