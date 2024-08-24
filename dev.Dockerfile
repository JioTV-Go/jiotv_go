FROM golang:1.23-alpine

ENV GO111MODULE=on \
    JIOTV_DEBUG=true

WORKDIR /app

RUN mkdir "/build"

# Copy source files from the host computer to the container
COPY . .

RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -build="go build -o build/jiotv_go ." -command="build/jiotv_go serve --public" -include="*.html"
