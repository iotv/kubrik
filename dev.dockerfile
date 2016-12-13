FROM golang:1.7

WORKDIR /opt
COPY ./ /opt/

RUN go build

ENTRYPOINT ./kubrik
