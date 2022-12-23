FROM golang:latest

COPY . /app
WORKDIR /app

RUN go build main.go

ENTRYPOINT ["./main"]