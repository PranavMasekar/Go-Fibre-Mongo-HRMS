FROM golang:1.16-alpine

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go mod download

RUN go build -o go-hrms .

EXPOSE 3000

CMD ["/app/go-hrms"]