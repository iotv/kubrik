FROM golang:1.7

WORKDIR /go
COPY ./ /go/src/github.com/mg4tv/kubrik

RUN go build github.com/mg4tv/kubrik

ENTRYPOINT ["./kubrik", "serve", "-p", "8080"]

