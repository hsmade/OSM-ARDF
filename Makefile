IMAGE := hsmade/osm-ardf
VERSION := $(shell git describe --tags 2>/dev/null)
BUILD := $(shell git rev-parse --short HEAD)

ifndef VERSION
VERSION = 0.0.0
endif

ifdef BUILD
TAG = "${VERSION}-${BUILD}"
else
TAG = "${VERSION}"
endif

image: test
	docker build -t ${IMAGE}:${TAG} -f build/package/Dockerfile .

lint:
	go fmt ./...
	go vet ./...

download:
	go mod tidy
	go mod download
	go mod verify

test: download lint
	go test -v ./...

compile:
	go build -ldflags="-w -s" -o dist/aprs_receiver ./cmd/aprs_receiver/aprs_receiver.go
	go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/udp_receiver/udp_receiver.go
	go build -ldflags="-w -s" -o dist/stdin_receiver ./cmd/stdin_receiver/stdin_receiver.go
	go build -ldflags="-w -s" -o dist/web_server ./cmd/web_server/web_server.go

clean:
	go clean

upload: image
	docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
	docker push ${IMAGE}:${TAG}
