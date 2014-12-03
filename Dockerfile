FROM golang

MAINTAINER Kristian Spriggs

ADD . /go/src/github.com/knspriggs/twitterweizen/

RUN go get github.com/tools/godep
RUN cd /go/src/github.com/knspriggs/twitterweizen && godep restore

RUN go install github.com/knspriggs/twitterweizen
ENTRYPOINT /go/bin/twitterweizen

EXPOSE 8080
