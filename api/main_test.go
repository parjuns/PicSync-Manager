package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

// const connStr = "user=root password=secret dbname=userdb sslmode=disable"

var testPostgresStore *PostgresStore
var testAPIServer *APIServer

func TestMain(m *testing.M) {

	setup()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}

func setup() {
	db, err := NewPostgresStore(&Config{
		"root", "secret", "user_db", "disable",
	})
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	testPostgresStore = db
	testAPIServer = NewAPIServer(":3000", testPostgresStore)
}

func teardown() {
	testPostgresStore.db.Exec("TRUNCATE TABLE products")
	testPostgresStore.db.Exec("ALTER TABLE products AUTO_INCREMENT = 1")
}

func makeRequest(method, url string, body []byte) *httptest.ResponseRecorder {
	request, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	writer := httptest.NewRecorder()
	writer.Header().Set("Content-Type", "application/json")
	router().ServeHTTP(writer, request)
	return writer
}

func router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/product", makeHTTPHandleFunc(testAPIServer.handleCreateProduct)).Methods("POST")
	router.HandleFunc("/product/{id}", makeHTTPHandleFunc(testAPIServer.handleGetProduct)).Methods("GET")
	router.HandleFunc("/product/{id}", makeHTTPHandleFunc(testAPIServer.handleUpdateProduct)).Methods("POST")
	return router
}
