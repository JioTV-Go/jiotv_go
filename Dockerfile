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

# Expose port 5001 to the outside world

EXPOSE 5001

# Command to run the executable

CMD ["./jiotv_go", ":5001"]
