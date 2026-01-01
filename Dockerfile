FROM reg.aetherial.dev/base-images/golang:1.25.5 as builder

RUN mkdir -p /tmp/build

WORKDIR /tmp/build

COPY ./Makefile .
COPY ./assets/ ./assets/
COPY ./cmd/ ./cmd/
COPY ./go.mod .
COPY ./go.sum .
COPY ./pkg/ ./pkg/
COPY ./scripts/ ./scripts/

RUN make build && make root-install

from reg.aetherial.dev/base-images/alpine:3.14 as final

COPY --from=builder /bin/keiji /bin/keiji
#### COPY --from=builder /bin/keiji-ctl /bin/keiji-ctl

CMD ["keiji", "-content", "embed"]

