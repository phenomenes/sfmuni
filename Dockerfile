FROM golang:1.7

MAINTAINER Jimena Cabrera Notari

RUN go get bitbucket.org/phenomenes/sfmuni

CMD [ "sfmuni", "--port=8080" ]
