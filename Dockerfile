FROM golang:1.18.2-alpine AS builder

WORKDIR /root

# add source code
ADD . .
# Dependencies and build
# Install dependencies, go, and cleanup
RUN go get \
    && go build -o banano-utils

# run main.go
CMD ["./banano-utils"]
