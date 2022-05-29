FROM golang:1.18.2-alpine

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN go get && go build -o banano-utils .

CMD ["/app/banano-utils"]
