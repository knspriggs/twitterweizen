FROM golang

MAINTAINER Kristian Spriggs

ADD . /app
WORKDIR /go
WORKDIR /

RUN go get ./...

RUN go install
CMD twitterweizen
