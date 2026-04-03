FROM golang:1.26.1 AS builder

WORKDIR /app

RUN apk add --no-cache make git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/bin/app .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./app"]
