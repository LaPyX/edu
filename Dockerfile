FROM golang:1.17

WORKDIR /opt/calls

RUN apt-get update \
    && apt-get install -y --no-install-recommends -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" curl

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

COPY ./go.mod .
COPY ./go.sum .

COPY . .

RUN go mod download

EXPOSE 8080

CMD ["air"]
