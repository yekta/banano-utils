FROM golang:1.18.2-bullseye AS build

WORKDIR /go/src/banano-utils

# Copy all the Code and stuff to compile everything
COPY . .

# Downloads all the dependencies in advance (could be left out, but it's more clear this way)
RUN go mod download

# Builds the application as a staticly linked one, to allow it to run on alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

# Moving the binary to the 'final Image' to make it smaller
FROM golang:1.18.2-alpine

WORKDIR /app

COPY --from=build /go/src/banano-utils/app .

CMD ["./app"]