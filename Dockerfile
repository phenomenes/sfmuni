FROM golang:1.7

MAINTAINER Jimena Cabrera Notari

RUN go get github.com/phenomenes/sfmuni

CMD [ "sfmuni", "--port=8080" ]
