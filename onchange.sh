#!/bin/bash

path=$1
dir=$(dirname "$path")
filename=$(basename "$path")
extension="${filename##*.}"
nameonly="${filename%.*}"

case $extension in
    go)
	gofmt -w $path
	;;
esac

GOPATH=$HOME
go generate ./...
go test -cover -coverprofile /tmp/c.out .
go tool cover -o /tmp/coverage.html -html /tmp/c.out
