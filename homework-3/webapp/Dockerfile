FROM golang:latest

WORKDIR /app

COPY . .

RUN go build main.go

CMD ["./main"]

EXPOSE 3000