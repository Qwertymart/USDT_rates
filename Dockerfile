FROM golang:1.26.1 AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y make git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/bin/app .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./app"]
