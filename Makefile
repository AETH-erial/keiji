.PHONY: build format test coverage dev-run install root-install


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

# Build as an unprivileged user first; invoke this target as root (sudo make install).
# Existing binaries and the vendor unit are replaced; configuration and data are not.
install:
	@test "$$(id -u)" -eq 0 || { echo "Run sudo make install after make build" >&2; exit 1; }
	@test -f ./build/linux/$(WEBSERVER)/$(WEBSERVER) -a -f ./build/linux/$(SEED_CMD)/$(SEED_CMD) || { echo "Run make build first" >&2; exit 1; }
	./deploy/setup.sh
	install -d -o root -g root -m 0755 /usr/local/bin /etc/systemd/system
	install -o root -g root -m 0755 ./build/linux/$(WEBSERVER)/$(WEBSERVER) /usr/local/bin/$(WEBSERVER)
	install -o root -g root -m 0755 ./build/linux/$(SEED_CMD)/$(SEED_CMD) /usr/local/bin/$(SEED_CMD)
	install -o root -g root -m 0644 ./deploy/keiji.service /etc/systemd/system/keiji.service
	@echo "Installed. Review /etc/keiji/keiji.env, then run systemctl daemon-reload and start keiji explicitly."


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
