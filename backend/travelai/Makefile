include .env

run: build

build:
	@go build -o build/travelai .

createdb:
	migrate -path db/migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@127.0.0.1:5432/${DB_NAME}?sslmode=disable" -verbose up

dropdb:
	migrate -path db/migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@127.0.0.1:5432/${DB_NAME}?sslmode=disable" -verbose down
