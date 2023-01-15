FROM golang:latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o main main.go 

EXPOSE 3005

ARG deploy_env

ENV DEPLOY_ENV=$deploy_env

ENTRYPOINT [ "./main" ]