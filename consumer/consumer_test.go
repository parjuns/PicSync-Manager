package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

const (
	test_url         = "http://localhost:3000/product"
	test_image_url   = "https://via.placeholder.com/100/1967505"
	test_path        = "./products_test.json"
	test_connDbStr   = "user=root password=secret dbname=userdb sslmode=disable"
	test_connAmqpStr = "amqp://guest:guest@localhost:5672/"
	test_queueName   = "testQueueService1"
	test_dirname     = "tests"
)

var testDb *sql.DB
var test_products = make([]string, 0)
var test_productIds = make([]string, 0)

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}
func setup() {
	db, err := sql.Open("postgres", test_connDbStr)
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	testDb = db

	var test_product1 = `{
		"name": "aged-thunder",
		"description": "black-dawn",
		"images": [
		  "https://via.placeholder.com/100/2225011",
		  "https://via.placeholder.com/100/378823"
		],
		"price": "501",
		"user_id": 19
	  }`
	var test_product2 = `{
		"name": "cold-butterfly",
		"description": "nameless-wind",
		"images": [
		  "https://via.placeholder.com/100/1300225",
		  "https://via.placeholder.com/100/1607819"
		],
		"price": "112",
		"user_id": 94
	  }`
	var test_product3 = `{
		"name": "autumn-sound",
		"description": "billowing-shadow",
		"images": [
		  "https://via.placeholder.com/100/2636711",
		  "https://via.placeholder.com/100/520910"
		],
		"price": "223",
		"user_id": 65
	  }`

	test_products = append(test_products, test_product1, test_product2, test_product3)

	for _, product := range test_products {
		test_productIds = append(test_productIds, createTestProduct([]byte(product)))
	}
}
func teardown() {
	testDb.Exec("TRUNCATE TABLE products")
	testDb.Exec("ALTER TABLE products AUTO_INCREMENT = 1")
	os.RemoveAll(test_dirname)
}

func Test_Consumer_ConnectAMQPReceiveMsg(t *testing.T) {

	createTestConnectAMQPSendMsg()

	err := doConsumeMsgWithTimeout()
	assert.NoError(t, err)

	// assert.EventuallyWithT(t, func(collect *assert.CollectT) {
	// 	connectAMQPReceiveMsg(test_connAmqpStr, test_queueName, test_url, test_dirname)
	// },
	// 	10*time.Second, 1*time.Second)
	// assert.
	// assert.Panics(t, func() { connectAMQPReceiveMsg(test_connAmqpStr, test_queueName, test_url, test_dirname) })
}
func doConsumeMsgWithTimeout() error {
	result := make(chan string, 1)
	go func() {
		connectAMQPReceiveMsg(test_connAmqpStr, test_queueName, test_url, test_dirname)
		result <- "done"
	}()
	select {
	case <-time.After(10 * time.Second):
		return nil
	case result := <-result:
		return errors.New("timeout" + result)
	}
}
func Test_Consumer_GetImageUrls(t *testing.T) {
	urls := getImageUrls(test_url, test_productIds[0])
	assert.Len(t, urls, 2)
	assert.Contains(t, urls, "https://via.placeholder.com/100/2225011")
	assert.Contains(t, urls, "https://via.placeholder.com/100/378823")
}

func Test_Consumer_DownloadStoreCompressImage(t *testing.T) {
	urls := []string{"https://via.placeholder.com/100/2225011", "https://via.placeholder.com/100/378823"}
	paths := downloadStoreCompressImage(urls, test_dirname, test_productIds[0])
	expectedpath1 := fmt.Sprintf("./%s/product_%s_img_%s.png", test_dirname, test_productIds[0], path.Base(urls[0]))
	expectedpath2 := fmt.Sprintf("./%s/product_%s_img_%s.png", test_dirname, test_productIds[0], path.Base(urls[1]))
	assert.Len(t, paths, 2)
	assert.Contains(t, paths, expectedpath1)
	assert.Contains(t, paths, expectedpath2)
}

func Test_Consumer_SetStoragePaths(t *testing.T) {
	urls := []string{"https://via.placeholder.com/100/2225011", "https://via.placeholder.com/100/378823"}
	path1 := fmt.Sprintf("./%s/product_%s_img_%s.png", test_dirname, test_productIds[0], path.Base(urls[0]))
	path2 := fmt.Sprintf("./%s/product_%s_img_%s.png", test_dirname, test_productIds[0], path.Base(urls[1]))
	paths := make([]string, 0)
	paths = append(paths, path1, path2)
	err := setStoragePaths(test_url, test_productIds[0], paths)
	assert.NoError(t, err)
}

func Test_Consumer_ImageProcessingWithCreateFolder(t *testing.T) {

	resp, err := http.Get(test_image_url)
	if err != nil {
		panic(err)
	}
	err = createFolder(test_dirname)
	assert.NoError(t, err)

	err = imageProcessing(resp.Body, test_dirname, "test_img_1")
	assert.NoError(t, err)
}

func Test_Consumer_FailOnError(t *testing.T) {
	msg := "msg for error"
	assert.NotPanics(t, func() { failOnError(nil, msg) })
}

func createTestProduct(payload []byte) string {
	r, err := http.NewRequest("POST", test_url, bytes.NewBuffer(payload))
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
	productMsg = strings.ReplaceAll(productMsg, "\"\n", "")
	productId := strings.Split(productMsg, ":")[1]
	return productId
}

func createTestConnectAMQPSendMsg() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := amqp.Dial(test_connAmqpStr)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		test_queueName, // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, id := range test_productIds {
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
		log.Printf(" [x] Test Sent Product with ID:%s\n", id)
	}
}
