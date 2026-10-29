.PHONY: build run test tidy lint

build:
	go build -o bin/trip-service cmd/trip-service/main.go

run: build
	./bin/trip-service

tidy:
	go mod tidy
