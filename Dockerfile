FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Install Logdy (log viewer)
ARG LOGDY_VERSION=0.17.1
ARG TARGETARCH
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then LOGDY_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then LOGDY_ARCH="arm64"; \
    else LOGDY_ARCH="$ARCH"; fi && \
    wget -q -O /usr/local/bin/logdy \
      "https://github.com/logdyhq/logdy-core/releases/download/v${LOGDY_VERSION}/logdy_linux_${LOGDY_ARCH}" && \
    chmod +x /usr/local/bin/logdy

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
