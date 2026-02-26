package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// Official RabbitMQ Go client
	amqp "github.com/rabbitmq/amqp091-go"
)

// Helper function for error handling (prints out error message and exits)
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	// 1. CONNECT TO RABBITMQ SERVER
	// Default credentials: guest/guest
	// Port 5672 is the default AMQP port
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close() // Always close connection when finished

	
	// 2. OPEN A CHANNEL
	// Most AMQP operations happen on channels, not directly on connection
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	
	// 3. DECLARE AN EXCHANGE
	// Exchange routes messages to queues
	// Type "direct" means routing is based on exact routing key match
	err = ch.ExchangeDeclare(
		"demo-exchange", // exchange name
		"direct",        // exchange type
		true,            // durable (survives broker restart)
		false,           // auto-delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // additional arguments
	)
	failOnError(err, "Failed to declare exchange")


	// 4. DECLARE A QUEUE
	// Queue stores messages until consumed
	q, err := ch.QueueDeclare(
		"demo-queue", // queue name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare queue")

	// 5. BIND QUEUE TO EXCHANGE
	// This connects the queue to the exchange using a routing key
	err = ch.QueueBind(
		q.Name,         // queue name
		"demo-key",     // routing key
		"demo-exchange", // exchange name
		false,
		nil,
	)
	failOnError(err, "Failed to bind queue")

	// 6. PUBLISH A MESSAGE
	// Create a timeout context for publishing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello from Go Producer!"

	err = ch.PublishWithContext(
		ctx,
		"demo-exchange", // exchange
		"demo-key",      // routing key (must match binding)
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent, // make message persistent
			Body:         []byte(body),    // message payload
		},
	)
	failOnError(err, "Failed to publish message")

	fmt.Println("Message Sent:", body)
}