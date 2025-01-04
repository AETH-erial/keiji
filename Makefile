.PHONY: build format docs


WEBSERVER = keiji
SEED_CMD = keiji-ctl
SWAG := $(shell command -v swag 2> /dev/null)
## Have to set the WEB_ROOT and DOMAIN_NAME environment variables when building
build:
	go build -o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	go build -o ./build/linux/$(SEED_CMD)/$(SEED_CMD) ./cmd/$(SEED_CMD)/$(SEED_CMD).go

install:
	sudo cp ./build/linux/$(SEED_CMD)/$(SEED_CMD) /usr/local/bin/

format:
	go fmt ./...

test:
	go test ./...


coverage:
	go test -v ./... -covermode=count -coverpkg=./... -coverprofile coverage/coverage.out
	go tool cover -html coverage/coverage.out -o coverage/coverage.html


dev-run:
	go build -o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	./build/linux/$(WEBSERVER)/$(WEBSERVER) .env
