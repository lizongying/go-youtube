.PHONY: youtubeServer youtubeSpider

all:  youtubeServer youtubeSpider


youtubeServer:
	go mod tidy
	go vet ./cmd/youtubeServer
	go build -ldflags "-s -w" -o  ./releases/youtubeServer  ./cmd/youtubeServer

youtubeSpider:
	go mod tidy
	go vet ./cmd/youtubeSpider
	go build -ldflags "-s -w" -o  ./releases/youtubeSpider  ./cmd/youtubeSpider