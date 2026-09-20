FROM golang:alpine AS builder

WORKDIR /app

# Install git since go mod tidy requires it to fetch certain dependencies
RUN apk add --no-cache git

# To avoid build issues if go.sum is missing
COPY go.mod ./
RUN go mod tidy || true

COPY . .
RUN go mod tidy

RUN go build -o /server ./cmd/server

# Final minimal stage
FROM alpine:latest

# Required for making HTTPS requests to Yahoo Finance, AWS, etc.
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled binary and necessary config files
COPY --from=builder /server /app/server
COPY --from=builder /app/config /app/config

CMD ["/app/server"]
