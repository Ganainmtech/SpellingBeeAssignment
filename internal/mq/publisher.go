package mq

import (
	"log"

	"github.com/streadway/amqp"
)

// Exchange for spelling bee events.
const exchangeName = "bee.events"

// Publisher holds the RabbitMQ connection and channel.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Connect opens a connection and channel and declares the topic exchange.
func Connect(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare a durable topic exchange for our events.
	if err := ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // args
	); err != nil {
		conn.Close()
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Publish sends a plain text message on the bee.events exchange
func (p *Publisher) Publish(routingKey, message string) {
	if p == nil || p.channel == nil {
		// Allow callers to treat nil publisher as "MQ disabled"
		return
	}

	err := p.channel.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Printf("RabbitMQ publish error: %v", err)
	}
}

// Close closes channel and connection.
func (p *Publisher) Close() {
	if p == nil {
		return
	}
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
