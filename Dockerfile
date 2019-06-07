# This file is a template, and might need editing before it works on your project.
FROM golang:1.12.5-alpine3.9 AS builder

# We'll likely need to add SSL root certificates
RUN apk add --no-cache ca-certificates git

RUN mkdir -p /go/src/github.com/dosko64/geoget/
WORKDIR /go/src/github.com/dosko64/geoget/

# COPY go.mod and go.sum files to the workspace
COPY go.mod /go/src/github.com/dosko64/geoget/
COPY go.sum /go/src/github.com/dosko64/geoget/

COPY . /go/src/github.com/dosko64/geoget/
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o geoget .

FROM alpine:3.9

# Since we started from scratch, we'll copy the SSL root certificates from the builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /server/

COPY data/ /server/data

COPY --from=builder /go/src/github.com/dosko64/geoget/geoget /server/geoget

EXPOSE 8000

ENTRYPOINT ["/server/geoget"]

