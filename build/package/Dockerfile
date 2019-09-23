FROM golang:1.13-alpine as builder
ADD . /go/src/github.com/hsmade/OSM-ARDF
WORKDIR /go/src/github.com/hsmade/OSM-ARDF
RUN adduser -D -g '' appuser
RUN apk update && apk add --no-cache make git ca-certificates && update-ca-certificates
RUN GOOS=linux GOARCH=amd64 make build

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/hsmade/OSM-ARDF/dist/* /
USER appuser
