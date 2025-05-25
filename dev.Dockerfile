FROM golang:1.24-alpine

ENV GO111MODULE=on \
    JIOTV_DEBUG=true \
    JIOTV_PATH_PREFIX="/app/.jiotv_go"

WORKDIR /app

RUN mkdir "/build"

# Copy source files from the host computer to the container
COPY . .

RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -build="go build -o build/jiotv_go ." -command="build/jiotv_go serve --public" -include="*.html"
