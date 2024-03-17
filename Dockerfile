FROM golang:1.22.1-alpine
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o main .
CMD ["./main"]
