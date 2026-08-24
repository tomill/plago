deploy: deps build
	doppler run -- lambroll deploy

deps:
	go mod tidy
	lambroll version
	doppler run -- true || doppler setup

build: rev=$(shell git rev-parse --short HEAD)
build:
	GOOS=linux GOARCH=amd64 go build -o plago -ldflags='-s -w -X github.com/tomill/plago/internal/config.version=$(rev)' ../cmd/plago/

invoke-test:
	echo '{"rawPath":"/plago","queryStringParameters":{"in":"dummy","out":"gmail"}}' | doppler run -- lambroll invoke --log-tail
