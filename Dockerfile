FROM alpine:latest

RUN apk add go chromium

WORKDIR /go/src/chrome-crawler
COPY . .

RUN go get -d -v ./...
RUN go build .

ENTRYPOINT ["./chrome-crawler"]
