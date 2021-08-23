FROM golang:1.17-alpine as builder

ENV GOPATH /

RUN mkdir /opt/yp && cd /opt/yp
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /opt/yp/yp

# alpine 3.10.3
FROM alpine:3.11
COPY --from=builder /opt/yp/yp /usr/bin/yp
ENTRYPOINT ["/usr/bin/yp"]
