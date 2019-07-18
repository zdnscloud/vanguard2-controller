VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"
GOSRC = $(shell find . -type f -name '*.go')

all: grpc $(GOSRC) 
	go build ${LDFLAGS} main.go

grpc: proto/dynamic_update_interface.proto proto/rrset.proto
	cd proto && protoc -I. --go_out=plugins=grpc:. *.proto

