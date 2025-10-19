FROM golang:1.24.5-alpine 

RUN apk add --no-cache git

WORKDIR /delayed-notifyer

COPY go.mod go.sum ./
COPY ./ ./

RUN go mod tidy

RUN go build -o delayed-notifyer ./cmd/main.go

CMD ["./delayed-notifyer"]