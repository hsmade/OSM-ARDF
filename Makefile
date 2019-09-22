VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)

package: test
	docker build -t hsmade/osm-ardf -f build/package/Dockerfile .

lint:
	go fmt ./...
	go vet ./...

download:
	@go mod tidy
	@go mod download
	@go mod verify

test: download lint
	go test -v ./...

build: test
	@go build -ldflags="-w -s" -o dist/aprs_receiver ./cmd/aprs_receiver/aprs_receiver.go
	@go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/udp_receiver/udp_receiver.go
	@go build -ldflags="-w -s" -o dist/web_server ./cmd/web_server/web_server.go

clean:
	go clean