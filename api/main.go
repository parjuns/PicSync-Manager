package main

import (
	"log"

	_ "github.com/lib/pq"
)

func main() {
	postgres, err := NewPostgresStore(&Config{
		"root", "secret", "user_db", "disable",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("db connection succesfull")

	server := NewAPIServer(":3000", postgres)
	server.Run()
}
