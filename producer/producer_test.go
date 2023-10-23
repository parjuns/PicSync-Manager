package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const (
	test_url         = "http://localhost:3000/product"
	test_path        = "./products_test.json"
	test_connDbStr   = "user=root password=secret dbname=userdb sslmode=disable"
	test_connAmqpStr = "amqp://guest:guest@localhost:5672/"
	test_queueName   = "testQueueService1"
)

var testDb *sql.DB

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
}
func teardown() {
	testDb.Exec("TRUNCATE TABLE products")
	testDb.Exec("ALTER TABLE products AUTO_INCREMENT = 1")
}

func Test_Producer_connectAMQPPublishWitMsg(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	test_productIds := createProducts(test_url, test_path)
	assert.NotPanics(t, func() { connectAMQPSendMsg(ctx, test_connAmqpStr, test_queueName, test_productIds) })
}

func Test_Producer_CreateProduct(t *testing.T) {
	var jsonStr = []byte(`{
		"name": "product1", 
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"price":"125",
		"user_id":17
	  }`)
	productId := createProduct(test_url, jsonStr)
	assert.NotZero(t, productId)
}

func Test_Producer_CreateProducts(t *testing.T) {
	productIds := createProducts(test_url, test_path)
	assert.NotEmpty(t, productIds)
	assert.Len(t, productIds, 3)
}

func Test_Producer_FailOnError(t *testing.T) {
	msg := "msg for error"
	assert.NotPanics(t, func() { failOnError(nil, msg) })
}
