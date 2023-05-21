FROM golang:1.16-alpine As builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o go-hrms .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/go-hrms /app/go-hrms

EXPOSE 3000

CMD ["/app/go-hrms"]