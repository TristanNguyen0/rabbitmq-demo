package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Helper function for consistent error handling
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	// 1. CONNECT TO RABBITMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// 2. OPEN A CHANNEL
	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()


	// ADDING THE QUEUE DECLARATION
	_, err = ch.QueueDeclare(
		"demo-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare queue")

	// 3. START CONSUMING MESSAGES
	msgs, err := ch.Consume(
		"demo-queue", // queue name
		"",           // consumer tag
		false,        // AUTO-ACK disabled (manual ACK required)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to register consumer")

	fmt.Println("Waiting for messages...")

	// 4. PROCESS MESSAGES
	// This loop continuously listens for messages
	for d := range msgs {

		fmt.Printf("Received message: %s\n", d.Body)

		// Simulate processing time
		time.Sleep(7 * time.Second)

		fmt.Println("Processing complete")

		// 5. MANUAL ACKNOWLEDGMENT
		// This tells RabbitMQ:
		// "The message was processed successfully."
		// If we DO NOT call Ack(), the message will be re-delivered.
		// err := d.Ack(false)
		// failOnError(err, "Failed to ACK message")
		// fmt.Println("ACK sent to RabbitMQ")
	}
}