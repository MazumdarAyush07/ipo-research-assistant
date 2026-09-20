FROM golang:alpine

WORKDIR /app

# To avoid build issues if go.sum is missing
COPY go.mod ./
RUN go mod tidy || true

COPY . .
RUN go mod tidy

RUN go build -o /server ./cmd/server
EXPOSE 8080
CMD ["/server"]
