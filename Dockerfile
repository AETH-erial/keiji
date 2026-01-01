FROM horus-harbor01.void-society.online/base-images/golang:1.25.5

RUN mkdir -p /tmp/build

WORKDIR /tmp/build

COPY ./Makefile .
COPY ./assets/ ./assets/
COPY ./cmd/ ./cmd/
COPY ./go.mod .
COPY ./go.sum .
COPY ./pkg/ ./pkg/
COPY ./scripts/ ./scripts/

RUN ls -lha && make build && make root-install

CMD ["keiji", "-content", "embed"]

