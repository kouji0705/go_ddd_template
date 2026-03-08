FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/app/main.go

FROM alpine:latest AS prod

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]
