FROM golang:alpine

WORKDIR /app

# Install git since go mod tidy requires it to fetch certain dependencies
RUN apk add --no-cache git

# To avoid build issues if go.sum is missing
COPY go.mod ./
RUN go mod tidy || true

COPY . .
RUN go mod tidy

RUN go build -o /server ./cmd/server
EXPOSE 8080
CMD ["/server"]
