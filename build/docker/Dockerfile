FROM golang:1.16.9-alpine3.14 AS builder

COPY go.mod go.sum /src/

WORKDIR /src/
RUN apk add git && \
    go mod download

COPY main.go /src/main.go
COPY internal /src/internal

RUN go build github.com/arivum/netpolviz

###

FROM alpine:3.14

COPY --from=builder /src/netpolviz /usr/local/bin/netpolviz
COPY assets/web /assets/web

ENTRYPOINT [ "/usr/local/bin/netpolviz" ]