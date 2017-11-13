FROM golang:1.9
EXPOSE 8000 2000 2001
WORKDIR /go/src/app
COPY . .



RUN go-wrapper download
RUN go-wrapper install