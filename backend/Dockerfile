FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o cyberhack-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/cyberhack-server .
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./cyberhack-server"]
