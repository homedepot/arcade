.PHONY: build test setup lint

build:
	CGO_ENABLED=0 go build -o arcade cmd/arcade/arcade.go

test:
	go test ./...

setup:
	go get -v -t -d ./...
	if [ -f Gopkg.toml ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
		dep ensure; \
	fi; \
	go mod tidy; \
	go mod verify; \

lint:
	golangci-lint run
