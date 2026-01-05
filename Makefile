.PHONY: run build test clean migrate

run:
	go run cmd/main.go

build:
	go build -o bin/srs-automation cmd/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ uploads/

migrate:
	go run cmd/main.go migrate

dev:
	air

install-deps:
	go mod download
	go mod tidy
