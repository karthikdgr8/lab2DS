# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./
COPY ./sem ./sem

RUN go build -o /http_server
RUN mkdir -p /app/files

EXPOSE 8080

CMD [ "/http_server","8080" ]
