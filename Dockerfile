FROM golang:1.24 AS builder

WORKDIR /app
COPY . ./

RUN apt-get update && apt-get install -y gcc binutils

WORKDIR /app/cmd/bot

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o budgetbot main.go \
    && strip budgetbot
