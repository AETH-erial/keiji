.PHONY: build format docs


WEBSERVER = webserver
SWAG := $(shell command -v swag 2> /dev/null)

build:
	go build -ldflags "-X main.WEB_ROOT=/home/aeth/keiji/html \
	-X main.DOMAIN_NAME=localhost \
	-X main.REDIS_ADDR=127.0.0.1 \
	-X main.REDIS_PORT=6666" \
	-o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go 

format:
	go fmt ./...

docs:
ifndef SWAG
	$(error "Could not find the swag binary.")
endif
	swag init -g ./cmd/$(WEBSERVER)/$(WEBSERVER).go

dev-run:
	go build -ldflags "-X main.WEB_ROOT=/home/aeth/keiji/html" \
	-o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	./build/linux/$(WEBSERVER)/$(WEBSERVER) .env