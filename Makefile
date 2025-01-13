build:
	@go build -o bin/snipshot ./cmd/web

run: build
	@./bin/snipshot

test: 
	@go test ./cmd/web -v
