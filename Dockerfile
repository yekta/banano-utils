FROM golang:1.18.2-alpine

WORKDIR /root

# add source code
ADD . .
# Dependencies and build
# Install dependencies, go, and cleanup
RUN apt-get update && apt-get install -y \
    gcc pkg-config \
    && go get \
    && go build -o banano-utils \
    && rm -rf /var/lib/apt/lists/*

# run main.go
CMD ["./banano-utils", "-host=0.0.0.0", "-port=3000", "-logtostderr"]
