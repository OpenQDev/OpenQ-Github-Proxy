FROM golang:1.20.6-alpine3.17

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main .

EXPOSE 3005

ARG deploy_env

ENV DEPLOY_ENV=$deploy_env

ENTRYPOINT [ "./main" ]
