FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy source files from the host computer to the container
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internals ./internals

# Build the Go app with optimizations
RUN go build -ldflags="-s -w" -trimpath -o /app/jiotv_go ./cmd/jiotv_go

# Stage 2: Create the final minimal image
# skipcq: DOK-DL3007
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go executable from the previous stage
COPY --from=builder /app/jiotv_go .

# Set environment variables
ENV JIOTV_CREDENTIALS_PATH=secrets

# Volume for credentials
VOLUME /app/secrets

# Expose port 5001 to the outside world
EXPOSE 5001

# Command to run the executable
CMD ["./jiotv_go", ":5001"]
