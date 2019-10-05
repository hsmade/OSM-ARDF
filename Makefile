IMAGE := hsmade/osm-ardf
VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)

ifndef VERSION
VERSION = 0.0.0
endif

ifdef BUILD
TAG = "${VERSION}-${BUILD}"
else
TAG = "${VERSION}"
endif

package: test
	docker build -t ${IMAGE}:${TAG} -f build/package/Dockerfile .

lint:
	go fmt ./...
	go vet ./...

download:
	@go mod tidy
	@go mod download
	@go mod verify

test: download lint
	$(eval CONTAINER = $(shell docker run -d -P -e POSTGRES_PASSWORD=postgres timescale/timescaledb-postgis:1.4.2-pg11))
	$(eval PORT = $(shell docker inspect $(CONTAINER) | grep HostPort | grep -E -o [0-9]+))
	@sleep 5
	@POSTGRES_PORT=$(PORT) go test -v ./... || (docker kill ${CONTAINER};false)
	@docker kill ${CONTAINER}

build:
	@go build -ldflags="-w -s" -o dist/aprs_receiver ./cmd/aprs_receiver/aprs_receiver.go
	@go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/udp_receiver/udp_receiver.go
	@go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/stdin_receiver/stdin_receiver.go
	@go build -ldflags="-w -s" -o dist/web_server ./cmd/web_server/web_server.go

clean:
	go clean

upload: package
	docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
	docker push ${IMAGE}:${TAG}
