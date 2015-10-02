FROM golang:1.4-cross

ENV GOPATH /go
ENV USER root

WORKDIR /go/src/github.com/SvenDowideit/doccheck

RUN go get github.com/russross/blackfriday

ADD . /go/src/github.com/SvenDowideit/doccheck
RUN go get -d -v
RUN go build -o doccheck doccheck.go

