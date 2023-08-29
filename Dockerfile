FROM golang:latest

# Set the Current Working Directory inside the container

WORKDIR /app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container

COPY . .

ENV GIN_MODE=release

# Download all the dependencies

RUN go mod download

# Build the Go app

RUN go build -o jiotv_go ./cmd/jiotv_go

# Remove all files and folderes except the executable

RUN find . -mindepth 1 -maxdepth 1 ! -name 'jiotv_go' ! -name '.jiotv_go' -exec rm -rf {} +

# Set credentials path
ENV JIOTV_CREDENTIALS_PATH=secrets

# Volume for credentials
VOLUME /app/secrets

# Expose port to the outside world

ARG PORT=5001
EXPOSE $PORT

# Command to run the executable

CMD ["./jiotv_go", $PORT]
