.PHONY: build format docs


WEBSERVER = webserver
SEED_CMD = seed
SWAG := $(shell command -v swag 2> /dev/null)
## Have to set the WEB_ROOT and DOMAIN_NAME environment variables when building
build:
	go build -ldflags " -X main.DOMAIN_NAME=$(DOMAIN_NAME)" \
	-o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go

format:
	go fmt ./...

docs:
ifndef SWAG
	$(error "Could not find the swag binary.")
endif
	swag init -g ./cmd/$(WEBSERVER)/$(WEBSERVER).go

build-seed-cmd:
	go build -o ./build/linux/$(SEED_CMD)/$(SEED_CMD) ./cmd/$(SEED_CMD)/$(SEED_CMD).go

dev-run:
	go build -ldflags "-X main.WEB_ROOT=$(WEB_ROOT)" \
	-o ./build/linux/$(WEBSERVER)/$(WEBSERVER) ./cmd/$(WEBSERVER)/$(WEBSERVER).go && \
	./build/linux/$(WEBSERVER)/$(WEBSERVER) .env
