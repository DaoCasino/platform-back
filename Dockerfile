FROM golang:1.13.4 AS builder
RUN go version
WORKDIR /app
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o app ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./app"]