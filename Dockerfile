WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o main main.go 

EXPOSE 3005

CMD [ "/main" ]