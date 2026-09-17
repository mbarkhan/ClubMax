# syntax=docker/dockerfile:1

FROM golang:1.27.0-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/clubmax-api ./main.go

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata postgresql17-client \
    && addgroup -S -g 10001 clubmax \
    && adduser -S -D -H -u 10001 -G clubmax clubmax \
    && mkdir -p /data/uploads /data/backups \
    && chown -R clubmax:clubmax /data
COPY --from=builder --chown=clubmax:clubmax /out/clubmax-api /usr/local/bin/clubmax-api
USER clubmax
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/clubmax-api"]
