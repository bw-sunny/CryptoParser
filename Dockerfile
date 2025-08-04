FROM golang:1.24.5-alpine AS builder

WORKDIR /app

EXPOSE :50052

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main server.go request_ticker.go

CMD ["./main"]



