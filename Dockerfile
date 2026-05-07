# syntax=docker/dockerfile:1.7

FROM golang:1.25.0-alpine3.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/main ./cmd/api
RUN GOBIN=/out go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/main /app/main
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY db/migrations /app/db/migrations
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /app/main /usr/local/bin/entrypoint.sh

EXPOSE 8080

ENV MIGRATIONS_DIR=/app/db/migrations

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
