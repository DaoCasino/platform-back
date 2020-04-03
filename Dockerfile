#FROM alpine:latest
#
#RUN apk --no-cache add ca-certificates
#WORKDIR /root/
#
#COPY ./.bin/app .
#COPY ./config/ ./config/

FROM golang:1.13.4 AS builder
RUN go version
WORKDIR /app
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o app ./cmd/main.go

FROM scratch
ARG port
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE $port
ENTRYPOINT ["./app"]