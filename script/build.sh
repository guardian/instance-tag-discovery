#!/bin/sh
set -e
build() {
  go build -o instance-tag-discovery-${GOOS}-${GOARCH} .
}
GOOS=linux GOARCH=amd64 build
GOOS=linux GOARCH=arm64 build
GOOS=darwin GOARCH=amd64 build
GOOS=darwin GOARCH=arm64 build
