FROM golang:1.20
RUN apt-get update
RUN apt-get -y install python3
RUN apt-get -y install python3-setuptools
RUN apt-get -y install python3-pip
RUN apt-get -y install python3-venv
#apt install python3.11-venv

RUN mkdir -p /app

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./app/go.mod ./app/go.sum ./
RUN go mod download && go mod verify

COPY ./app .

RUN python3 -m venv .venv
RUN source .venv/bin/activate
RUN pip install -r ./requirements.txt

RUN go build -v -o app

CMD ["./app"]
