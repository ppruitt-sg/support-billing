FROM golang:1.10.3

WORKDIR /src
COPY . /src

RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/schema
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
RUN go build -o eb-go-app

EXPOSE 8080
CMD ["./eb-go-app"]
