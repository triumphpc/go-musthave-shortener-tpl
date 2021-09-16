FROM golang:latest

ADD . /app/
WORKDIR /app

COPY . .

RUN go build -o main cmd/shortener/main.go
CMD ["./main"]
