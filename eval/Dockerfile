FROM golang:1.25.6 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/eval
RUN go build -o eval

FROM ubuntu:latest
WORKDIR /app

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /app/eval/eval .
COPY eval/queries.jsonl ./

CMD ["./eval"]
