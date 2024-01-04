.PHONY: build test setup lint

build:
	CGO_ENABLED=0 go build -o arcade cmd/arcade/arcade.go

test:
	go generate ./...
	go test ./...

setup:
	go get -v -t -d ./...
	if [ -f go.mod ]; then \
		go mod tidy; \
		go mod verify; \
	fi; \
	if [ -f Gopkg.toml ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
		dep ensure; \
	fi; \

lint:
	golangci-lint run --skip-files .*_test.go --enable wsl --enable misspell --timeout 180s
