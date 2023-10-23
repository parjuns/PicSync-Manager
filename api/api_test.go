package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_API_HandleCreateProduct(t *testing.T) {

	var body = []byte("")
	writer := makeRequest("POST", "/product", body)
	msg := writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "")

	var jsonStr1 = []byte(`{
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"user_id":17
	  }`)
	writer = makeRequest("POST", "/product", jsonStr1)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "missing fields")

	var jsonStr2 = []byte(`{
		"name": "product1", 
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"price":"125",
		"user_id":1000
	  }`)
	writer = makeRequest("POST", "/product", jsonStr2)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "user id not found")

	var jsonStr3 = []byte(`{
		"name": "product1", 
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"price":"125",
		"user_id":17
	  }`)
	writer = makeRequest("POST", "/product", jsonStr3)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusCreated, writer.Code)
	assert.Contains(t, msg, "product added successfully with product id:")

}

func Test_API_HandleGetProduct(t *testing.T) {
	writer := makeRequest("GET", "/product/"+"abcd", nil)
	msg := writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "bad product id")

	writer = makeRequest("GET", "/product/"+"10000", nil)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "product id not found")

	var jsonStr1 = []byte(`{
		"name": "product1", 
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"price":"125",
		"user_id":17
	  }`)
	writer = makeRequest("POST", "/product", jsonStr1)
	msg = writer.Body.String()
	productId := strings.Split(strings.ReplaceAll(msg, "\"\n", ""), ":")[1]

	writer = makeRequest("GET", "/product/"+productId, nil)

	assert.Equal(t, http.StatusOK, writer.Code)

}

func Test_API_HandleUpdateProduct(t *testing.T) {
	writer := makeRequest("POST", "/product/"+"abcd", nil)
	msg := writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "bad product id")

	var jsonStr1 = []byte(`{
		"name": "product1", 
		"description": "this is product 1",
		"images":["https://via.placeholder.com/100/13234","https://via.placeholder.com/100/5675463"],
		"price":"125",
		"user_id":17
	  }`)
	writer = makeRequest("POST", "/product", jsonStr1)
	msg = writer.Body.String()
	productId := strings.Split(strings.ReplaceAll(msg, "\"\n", ""), ":")[1]

	var jsonStr2 = []byte("")
	writer = makeRequest("POST", "/product/"+productId, jsonStr2)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Contains(t, msg, "payload decode error")

	var jsonStr3 = []byte(`[
		"./home/path1",
		"./home/path2"
	  ]`)
	writer = makeRequest("POST", "/product/"+productId, jsonStr3)
	msg = writer.Body.String()
	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Contains(t, msg, `"compressed_images":["./home/path1","./home/path2"]`)

}
