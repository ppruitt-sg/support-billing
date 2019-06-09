FROM golang:1.12.5

#Install dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/github.com/ppruitt-sg/support-billing
COPY . ./

###### Commenting this out since dep handles the dependencies #####
# RUN go get github.com/go-sql-driver/mysql
# RUN go get github.com/gorilla/schema
# RUN go get github.com/gorilla/mux
# RUN go get github.com/gorilla/handlers
# RUN go get github.com/stretchr/testify/assert
# RUN go get github.com/stretchr/testify/require
# RUN go get github.com/ppruitt-sg/support-billing
# RUN go get github.com/kelseyhightower/envconfig
###################################################################

RUN dep ensure

RUN go build -o eb-go-app

EXPOSE 8080
CMD ["./eb-go-app"]
