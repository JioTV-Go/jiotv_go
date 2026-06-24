FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOEXPERIMENT=jsonv2,greenteagc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/jiotv_go .

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/jiotv_go .

ENV JIOTV_PATH_PREFIX=/app/.jiotv_go

VOLUME ["/app/.jiotv_go"]

EXPOSE 5001

ENTRYPOINT ["./jiotv_go"]
CMD ["--skip-update-check","serve","--public"]
