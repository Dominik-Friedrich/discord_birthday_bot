FROM golang:1.20

RUN mkdir -p /app

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./app/go.mod ./app/go.sum ./
RUN go mod download && go mod verify

COPY ./app .
RUN go build -v -o app

CMD ["./app"]
