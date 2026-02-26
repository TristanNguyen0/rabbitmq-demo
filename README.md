# RabbitMQ Live Demo
This demo will be run on Windows using WSL, but the same commands can be followed on Linux directly, and on macOS with the only difference being that installations (Go and Docker) should be done using Homebrew.

## Prerequisites
1. WSL (Windows Subsystem for Linux)
2. Docker
3. Go

### Install WSL
Open PowerShell as Admin: 
```PowerShell
wsl --install
```
Restart computer to finish WSL installation.

### Install [Docker Desktop](https://www.docker.com/products/docker-desktop/)
Visit the link above and download and install Docker Desktop.
#### Integrate docker into WSL:
1. Open Docker Desktop
2. Go to Settings → Resources → WSL Integration
3. Enable integration for your Ubuntu distribution
4. Click Apply & Restart
5. Verify docker installation in WSL:
```bash
docker --version
```

#### Docker Installation *for Linux*: (from official [Docker docs](https://docs.docker.com/engine/install/ubuntu/))
1. Set up Docker's apt repository
```bash
# Add Docker's official GPG key:
sudo apt update
sudo apt install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
```

2. Install the Docker packages
```bash
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```
3. Enable Docker without sudo
```bash
sudo usermod -aG docker $USER
newgrp docker
```
4. Verify Installation
```
bash
docker --version
```

### Install Go 
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
```
Add Go to PATH:
```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify Installation:
```bash
go version
```

## Getting Started with RabbitMQ
### Start RabbitMQ using Docker:
```bash
docker run -d --name rabbitmq-demo \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```
Ports:

5672 → AMQP

15672 → Web UI

Access Web UI at: http://localhost:15672

default user: guest

default password: guest



### Create Go Project 
We will be using the [RabbitMQ Advanced Message Queueing Protocol (AMQP) Go Client Library](https://github.com/rabbitmq/amqp091-go)
```bash
mkdir rabbitmq-demo
cd rabbitmq-demo
go mod init rabbitmq-demo
go get github.com/rabbitmq/amqp091-go
```

### Create producer.go and consumer.go files:
The architecture for this demo will be as follows:
Producer → Exchange → Queue → Consumer
```bash
touch producer.go consumer.go
```

### producer.go
#### Basic template for our go file:
Includes importing RabbitMQ Go Client, and creating a helper function for handling errors
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
}
```

#### Connect to RabbitMQ server and Open a channel
```go
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
```

#### Declaring Exchanges and Queues (and binding queue to exchange)
```go
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
	nil,             // no additional arguments
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
	nil,          // no additional arguments
)
failOnError(err, "Failed to declare queue")

// 5. BIND QUEUE TO EXCHANGE
// This connects the queue to the exchange using a routing key
err = ch.QueueBind(
	q.Name,         // queue name
	"demo-key",     // routing key
	"demo-exchange", // exchange name
	false, // no-wait
	nil, // no additional arguments
)
failOnError(err, "Failed to bind queue")
```

#### Publishing Message:
```go
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
```
timeout context ensures that if something goes wrong, the program does not block forever. The timeout is set to 5 seconds

### consumer.go
#### Basic template for our go file:
Includes importing RabbitMQ Go Client, and creating a helper function for handling errors
```go
package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
}
```
#### Connect to RabbitMQ and open channel:
```go
// 1. CONNECT TO RABBITMQ
conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
failOnError(err, "Failed to connect to RabbitMQ")
defer conn.Close()

// 2. OPEN A CHANNEL
ch, err := conn.Channel()
failOnError(err, "Failed to open channel")
defer ch.Close()
```

#### Consuming messages with ACK handling
```go
// 3. START CONSUMING MESSAGES
msgs, err := ch.Consume(
	"demo-queue", // queue name
	"",           // consumer tag
	false,        // AUTO-ACK disabled (manual ACK required)
	false,        // exclusive
	false,        // no-local
	false,        // no-wait
	nil,          // no additional arguments
)
failOnError(err, "Failed to register consumer")

fmt.Println("Waiting for messages...")
```

#### Processing messages (manual ACK)
```go
// 4. PROCESS MESSAGES
// This loop continuously listens for messages
for d := range msgs {

	fmt.Printf("Received message: %s\n", d.Body)

	// Simulate processing time
	time.Sleep(5 * time.Second)

	fmt.Println("Processing complete")

	// 5. MANUAL ACKNOWLEDGMENT
	// This tells RabbitMQ:
	// "The message was processed successfully."
	// If we DO NOT call Ack(), the message will be re-delivered.
	err := d.Ack(false)
	failOnError(err, "Failed to ACK message")

	fmt.Println("ACK sent to RabbitMQ")
}
```

## Missing Queue Declaration in consumer.go
### Stopping and removing RabbitMQ container to clear the declared queues:
```bash
docker stop rabbitmq-demo
docker rm rabbitmq-demo
```
### Restarting rabbitmq docker:
```bash
docker run -d --name rabbitmq-demo \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

### When trying to run consumer.go before producer.go:
```bash
Exception (404) Reason: "NOT_FOUND - no queue 'demo-queue' in vhost '/'"
```
This is because RabbitMQ does not auto-create queues on consume, **both producer and consumer should declare the queue.**
Queue declaration is idempotent, meaning if the queue already exists nothing happens, and if the queue doesn't exist, it is created.

Ultimately, declaring a queue twice is safe.

### Declaring queue in consumer.go:
```go
	// ADDING THE QUEUE DECLARATION (ensure this is before calling Consume)
	_, err = ch.QueueDeclare(
		"demo-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare queue")
```
declaring the queue in both producer.go and consumer.go ensures


## Demonstrating guaranteed delivery
### Comment out ACK from consumer.go:
```go
// err := d.Ack(false)
// failOnError(err, "Failed to ACK message")
// fmt.Println("ACK sent to RabbitMQ")
``` 
### Run consumer.go:
```bash
go run consumer.go
```
### Run producer.go: (in separate terminal)
```bash
go run producer.go
```
use ctrl+c to kill consumer and run consumer.go again:
```bash
go run consumer.go
```
This demonstrates guaranteed delivery

