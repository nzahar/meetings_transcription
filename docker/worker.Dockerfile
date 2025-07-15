FROM golang:1.22-alpine

WORKDIR /app

COPY worker/ ./worker/
WORKDIR /app/worker

RUN go mod download
RUN go build -o worker .

CMD ["./worker"]