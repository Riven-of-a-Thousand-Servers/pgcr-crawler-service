FROM golang:1.23 as builder

WORKDIR /app
COPY . .

RUN go mod tidy && GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -tags=prod -o pgcr-crawler-service ./cmd/crawler/

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/pgcr-crawler-service .

CMD ["/root/pgcr-crawler-service", "-workers=100", "-pgcr=15000000000", "-passes=1", "-env=prod"]
