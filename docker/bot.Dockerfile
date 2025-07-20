FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.work ./
COPY bot/ ./bot/
COPY worker/ ./worker/
COPY shared/ ./shared/

WORKDIR /app/bot

RUN go mod download
RUN go build -o bot ./cmd/bot

CMD ["./bot"]