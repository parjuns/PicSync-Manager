package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connAmqpStr = "amqp://guest:guest@localhost:5672/"
const productsLocPath = "./products.json"
const apiurl = "http://localhost:3000/product"
const queueName = "QueueService1"

func main() {
	productIds := createProducts(apiurl, productsLocPath)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connectAMQPSendMsg(ctx, connAmqpStr, queueName, productIds)
}

func connectAMQPSendMsg(ctx context.Context, connect, queueName string, productIds []string) {
	conn, err := amqp.Dial(connect)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, id := range productIds {
		err := ch.PublishWithContext(ctx,
			"",         // exchange
			queue.Name, // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(id),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent Product with ID:%s\n", id)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func createProducts(url, path string) (productIds []string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	body, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var products []interface{}

	if err := json.Unmarshal(body, &products); err != nil {
		panic(err)
	}
	productIds = make([]string, 0)
	for _, product := range products {
		byte, err := json.Marshal(product)
		if err != nil {
			panic(err)
		}
		id := createProduct(url, byte)
		productIds = append(productIds, id)
	}

	return productIds
}

func createProduct(url string, payload []byte) (productId string) {
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	productMsg := string(body)
	productMsg = strings.ReplaceAll(productMsg, "\"", "")
	productId = strings.Split(productMsg, ":")[1]
	return productId
}
