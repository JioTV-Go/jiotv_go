FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy source files from the host computer to the container
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg
COPY web ./web
COPY internal ./internal
COPY main.go ./main.go
COPY VERSION ./VERSION

# Build the Go app with optimizations
RUN go build -ldflags="-s -w" -trimpath -o /app/jiotv_go .

# Stage 2: Create the final minimal image
# skipcq: DOK-DL3007
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go executable from the previous stage
COPY --from=builder /app/jiotv_go .

# Set environment variables
ENV JIOTV_PATH_PREFIX="/app/.jiotv_go"

# Volume for credentials
VOLUME /app/.jiotv_go

# Expose port 5001 to the outside world
EXPOSE 5001

# Command to run the executable with arguments
# The CMD instruction has been replaced with ENTRYPOINT to allow arguments
ENTRYPOINT ["./jiotv_go"]

# Default arguments
CMD ["--skip-update-check", "serve", "--public"]
