FROM golang:1.20

# Install Python and its dependencies
RUN apt-get update && apt-get -y install python3 python3-setuptools python3-pip

# Create the working directory
RUN mkdir -p /app
WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./app/go.mod ./app/go.sum ./
RUN go mod download && go mod verify

COPY ./app .

# Install Python dependencies
COPY ./app/requirements.txt ./
RUN rm /usr/lib/python3.11/EXTERNALLY-MANAGED
RUN pip3 install -r ./requirements.txt

# Build the Go application
RUN go build -v -o app

CMD ["./app"]