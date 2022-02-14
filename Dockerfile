FROM golang:1.17.4-alpine3.15 AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

ENV GOARCH=amd64
RUN go build -ldflags "-X main.buildcommit=`git rev-parse --short HEAD` \ 
    -X main.buildtime=`date "+%Y-%m-%dT%H:%M:%S%Z:00"`" \
    -o /go/bin/app 

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
COPY /wwwroot /wwwroot
ENTRYPOINT /app
LABEL Name=gotodos Version=0.0.1
EXPOSE 8080