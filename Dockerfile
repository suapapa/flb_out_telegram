FROM golang:latest AS builder

WORKDIR /src

COPY . .
RUN go build -buildmode=c-shared -o out_telegram.so *.go

# ---
FROM fluent/fluent-bit:latest-debug

COPY --from=builder /src/out_telegram.so /plugins/out_telegram.so
COPY ./conf /conf

CMD ["/fluent-bit/bin/fluent-bit", "-c", "/conf/flb.conf", "-e", "/plugins/out_telegram.so"]
