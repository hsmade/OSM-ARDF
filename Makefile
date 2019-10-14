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
	$(eval CONTAINER := $(shell docker run -d -P -e POSTGRES_PASSWORD=postgres timescale/timescaledb-postgis:1.4.2-pg11))
	@echo waiting for postgres to startup
	@sleep 5
	$(eval PORT := $(shell docker inspect $(CONTAINER) | grep HostPort | grep -E -o [0-9]+))
	$(eval IP := $(shell docker inspect $(CONTAINER) | grep HostIp| grep -E -o [0-9.]+))
	@echo init postgres
	@docker exec -ti -e PGPASSWORD=postgres $(CONTAINER) psql -U postgres -c "create table doppler (time TIMESTAMP not null default now(), name text not null, lat real not null, lon real not null, bearing INT);"
	@docker exec -ti -e PGPASSWORD=postgres $(CONTAINER) psql -U postgres -c "SELECT create_hypertable('doppler', 'time', chunk_time_interval => interval '1 minute');"
	@echo postgress: $(IP):$(PORT)
	POSTGRES_PORT=$(PORT) POSTGRES_IP=$(IP) go test -v ./... || (docker kill ${CONTAINER};false)
	@docker kill ${CONTAINER}

compile:
	@go build -ldflags="-w -s" -o dist/aprs_receiver ./cmd/aprs_receiver/aprs_receiver.go
	@go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/udp_receiver/udp_receiver.go
	@go build -ldflags="-w -s" -o dist/udp_receiver ./cmd/stdin_receiver/stdin_receiver.go
	@go build -ldflags="-w -s" -o dist/web_server ./cmd/web_server/web_server.go

clean:
	go clean

upload: image
	docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
	docker push ${IMAGE}:${TAG}
