FROM golang:1.12.5

WORKDIR /src
COPY . /src

RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/schema
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
RUN go get github.com/stretchr/testify/assert
RUN go get github.com/stretchr/testify/require
RUN go get github.com/ppruitt-sg/support-billing
RUN go get github.com/kelseyhightower/envconfig

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

RUN go build -o eb-go-app

EXPOSE 8080
CMD ["./eb-go-app"]
