FROM golang

MAINTAINER Kristian Spriggs

ADD . /go/src/github.com/knspriggs/twitterweizen/

RUN go get github.com/knspriggs/twitterweizen

RUN go install github.com/knspriggs/twitterweizen
CMD /go/bin/twitterweizen

EXPOSE 8080
