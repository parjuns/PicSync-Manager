package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	compression "github.com/nurlantulemisov/imagecompression"
	amqp "github.com/rabbitmq/amqp091-go"
)

const baseUrl = "http://localhost:3000/product"
const connAmqpStr = "amqp://guest:guest@localhost:5672/"
const queueName = "QueueService1"
const dirname = "images"

func main() {
	connectAMQPReceiveMsg(connAmqpStr, queueName, baseUrl, dirname)
}

func connectAMQPReceiveMsg(connect, queueName, baseUrl, dirname string) {
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

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	go func() {
		for data := range msgs {
			//Extract productId from msg and log
			productId := strings.ReplaceAll(string(data.Body), "\n", "")
			log.Printf("Received a message: ProductID:%s added", productId)

			//Get imageurls using productId
			imageUrls := getImageUrls(baseUrl, productId)

			//Download images,compress them and store them
			storagePaths := downloadStoreCompressImage(imageUrls, dirname, productId)

			//Set paths on Database using Api
			if err := setStoragePaths(baseUrl, productId, storagePaths); err != nil {
				panic(err)
			}

			//log paths
			for _, path := range storagePaths {
				log.Printf("ProductID:%s ImagePath:%s added", productId, path)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func getImageUrls(baseUrl, productId string) []string {

	url := fmt.Sprintf("%s/%s", baseUrl, productId)

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

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

	var product map[string]interface{}
	if err := json.Unmarshal(body, &product); err != nil {
		panic(err)
	}

	imageUrls := make([]string, 0)
	for _, url := range product["images"].([]interface{}) {
		imageUrls = append(imageUrls, url.(string))
	}

	return imageUrls
}

func downloadStoreCompressImage(urls []string, dirname string, productId string) []string {
	paths := make([]string, 0)
	for _, url := range urls {
		r, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		if err := createFolder(dirname); err != nil {
			panic(err)
		}

		fname := fmt.Sprintf("product_%s_img_%s.png", productId, path.Base(url))
		if err := imageProcessing(r.Body, dirname, fname); err != nil {
			panic(err)
		}

		path := fmt.Sprintf("./%s/%s", dirname, fname)

		paths = append(paths, path)

	}
	return paths
}

func setStoragePaths(baseUrl, productId string, paths []string) error {
	url := fmt.Sprintf("%s/%s", baseUrl, productId)
	payload, err := json.Marshal(paths)
	if err != nil {
		return err
	}
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		return err
	}

	return nil
}

func createFolder(dirname string) error {
	_, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirname, 0755)
		if errDir != nil {
			return errDir
		}
	}
	return nil
}

func imageProcessing(body io.Reader, dirname, filename string) error {

	file, err := os.Create("./" + dirname + "/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := png.Decode(body)
	if err != nil {
		return err
	}
	compressing, _ := compression.New(90)
	compressingImage := compressing.Compress(img)

	if err := png.Encode(file, compressingImage); err != nil {
		return err
	}
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
