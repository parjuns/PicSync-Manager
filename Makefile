postgres:
	docker run --name mypostgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres 

createdb:
	docker exec -it mypostgres createdb --username=root --owner=root user_db

dropdb:
	docker exec -it mypostgres dropdb user_db

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/user_db?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/user_db?sslmode=disable" -verbose down

rabbitmq:
	docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.12-management

test:
	go test -v -cover ./...

.PHONY:postgres createdb dropdb migrateup migratedown rabbitmq test	