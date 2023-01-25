FROM golang:latest AS builder

ARG PLUGIN_VERSION=dev

WORKDIR /src

COPY . .
RUN go build \
    -buildmode=c-shared \
    -ldflags "-X main.pluginVersion=${PLUGIN_VERSION}" \
    -o out_telegram.so *.go

# ---
FROM fluent/fluent-bit:latest

COPY --from=builder /src/out_telegram.so /plugins/out_telegram.so
COPY ./conf /conf

CMD ["/fluent-bit/bin/fluent-bit", "-c", "/conf/flb.conf"]
