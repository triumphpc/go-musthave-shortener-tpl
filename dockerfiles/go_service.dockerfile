FROM golang:latest

# create workdir
ADD . /app/
WORKDIR /app

# copy all file from to workdir
COPY . .

# instal psql
RUN apt-get update
RUN apt-get -y install postgresql-client

RUN echo $PATH

# build go app
RUN go mod download
RUN go build -o main cmd/shortener/main.go
CMD ["./main"]
