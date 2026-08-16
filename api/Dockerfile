FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /kylistran-api ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /kylistran-api ./kylistran-api
EXPOSE 8080
ENTRYPOINT ["./kylistran-api"]
