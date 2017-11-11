FROM golang:1.9

WORKDIR /go/src/app
COPY . .

EXPOSE 8000 2000 2001

RUN go-wrapper download
RUN go-wrapper install