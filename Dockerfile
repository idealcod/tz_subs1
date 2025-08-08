FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12
RUN swag init

RUN go build -o main .
RUN chmod +x ./main

CMD ["/app/main"]
