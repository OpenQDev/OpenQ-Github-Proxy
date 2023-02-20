FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main ./src

EXPOSE 3005

ARG deploy_env

ENV DEPLOY_ENV=$deploy_env

ENTRYPOINT [ "./main" ]