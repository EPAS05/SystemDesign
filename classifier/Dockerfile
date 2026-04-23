FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/cli
RUN CGO_ENABLED=0 GOOS=linux go build -o classifier .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/cmd/cli/classifier .

COPY --from=builder /app/sql ./sql

CMD ["./classifier"]