FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.work ./
COPY bot/ ./bot/
COPY worker/ ./worker/
COPY shared/ ./shared/

WORKDIR /app/worker

RUN go mod download
RUN go build -o worker ./cmd/worker

CMD ["./worker"]