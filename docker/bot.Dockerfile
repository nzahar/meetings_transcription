FROM golang:1.24.5-alpine

WORKDIR /app

COPY bot/ ./bot/
WORKDIR /app/bot

RUN go mod download
RUN go build -o bot .

CMD ["./bot"]