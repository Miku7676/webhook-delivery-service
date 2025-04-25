FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app/

RUN go build -o webhook-service /app/

EXPOSE 8080
CMD ["./webhook-service"]