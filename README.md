# GO-Message-Queue-Api

This is go lang backend application in which user can perform CRUD operations on API Server using Producer and Consumer with RabbitMQ

# Requirements

- Go Language
- Text editor

# How to run

Steps to run the code

## Setup Postgres Database

```
make postgres
make createdb
make migrateup
```

## Start server on http://localhost:3000/

```
cd api
go build -o api.exe
.\api
```

## Start RabbitMQ server on http://localhost:5672/

```
make rabbitmq
```

## Now run Producer service on another Terminal

```
cd producer
go run producer.go
```

## Similarly run Consumer service on another Terminal

```
cd consumer
go run consumer.go
```

Go to http://localhost:15672 for RabbitMQ dashboard

Go to `./consumer/images` directory for downloaded images.

## Run Test and Coverage

```
make test
```
