FROM golang:1.24.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/ingest
RUN go build -o ingest

FROM ubuntu:latest
WORKDIR /app

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /app/ingest/ingest .
COPY ingest/data/ ./ingest/data/

CMD ["./ingest"]
