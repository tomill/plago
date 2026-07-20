deploy: deps build
	doppler run -- lambroll deploy

deps:
	go mod tidy
	lambroll version
	doppler run -- true || doppler setup

build: rev=$(shell git rev-parse --short HEAD)
build:
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w -X github.com/tomill/plago/config.version=$(rev)' ../
