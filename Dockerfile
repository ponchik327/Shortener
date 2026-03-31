# Stage 1: build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shortener ./cmd/app/main.go

# Stage 2: runtime
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/shortener .
COPY migrations/ migrations/
COPY templates/  templates/
COPY config.yaml .
EXPOSE 8080
ENTRYPOINT ["./shortener"]
CMD ["--config", "config.yaml"]
