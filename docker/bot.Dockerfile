FROM golang:1.22-alpine

WORKDIR /app

COPY bot/ ./bot/
WORKDIR /app/bot

RUN go mod download
RUN go build -o bot .

CMD ["./bot"]