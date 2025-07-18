FROM golang:1.24.5-alpine

WORKDIR /app

COPY worker/ ./worker/
WORKDIR /app/worker

RUN go mod download
RUN go build -o worker ./cmd/worker

CMD ["./worker"]