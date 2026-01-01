.PHONY: build format test coverage dev-run install


WEBSERVER = keiji
SEED_CMD = keiji-ctl
SWAG := $(shell command -v swag 2> /dev/null)
## Have to set the WEB_ROOT and DOMAIN_NAME environment variables when building
build:
	mkdir -p ./build/linux/$(WEBSERVER) ./build/linux/$(SEED_CMD)
	go build -o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	go build -o ./build/linux/$(SEED_CMD)/$(SEED_CMD) ./cmd/$(SEED_CMD)/$(SEED_CMD).go

root-install:
	cp ./build/linux/$(SEED_CMD)/$(SEED_CMD) /bin \
	&& cp ./build/linux/$(WEBSERVER)/$(WEBSERVER) /bin

install:
	sudo mkdir -p /usr/local/bin \
	&& sudo cp ./build/linux/$(SEED_CMD)/$(SEED_CMD) /usr/local/bin/ \
	&& sudo cp ./build/linux/$(WEBSERVER)/$(WEBSERVER) /usr/local/bin/

format:
	go fmt ./...

test:
	go test -v ./...


coverage-html:
	mkdir -p coverage/
	go test -v ./... -covermode=count -coverpkg=./... -coverprofile coverage/coverage.out
	go tool cover -html coverage/coverage.out -o coverage/coverage.html

coverage:
	go test ./... -cover


dev-run:
	go build -o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	./build/linux/$(WEBSERVER)/$(WEBSERVER) .env
