FROM golang:1.12.5-alpine3.9 AS build

RUN mkdir -p /go/src/github.com/zdnscloud/vanguard2-controller
COPY . /go/src/github.com/zdnscloud/vanguard2-controller

WORKDIR /go/src/github.com/zdnscloud/vanguard2-controller
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/github.com/zdnscloud/vanguard2-controller/vanguard2-controller /usr/local/bin/

ENTRYPOINT ["vanguard2-controller"]
