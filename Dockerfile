FROM golang:1.9
# This specifies the container executable, which in this case is "app".
# Don't change this - app is created when the Dockerfile does RUN go build
ENTRYPOINT [ "/app/app" ]
EXPOSE 2000 2001
WORKDIR /app
# Set GOPATH so go build doesn't lose its shit
ENV GOPATH /app 
COPY . .
RUN git clone https://github.com/anteater2/bitmesh.git /app/src/github.com/anteater2/bitmesh
RUN go build