FROM golang:1.18.2-alpine

RUN mkdir /app

WORKDIR /app

COPY . /app

RUN go get && go build -o banano-utils .

CMD ["/app/banano-utils"]
