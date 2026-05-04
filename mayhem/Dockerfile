FROM golang:1.25 as builder

COPY . /ocpp-go
WORKDIR /ocpp-go

RUN go install github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest
RUN apt update && apt install -y clang

RUN go get github.com/dvyukov/go-fuzz/go-fuzz-dep && \
    go-fuzz-build -libfuzzer -o /ocpp-go/mayhem/fuzzOcpp.a ./mayhem && \
    clang -fsanitize=fuzzer /ocpp-go/mayhem/fuzzOcpp.a -o /ocpp-go/mayhem/fuzzOcpp.libfuzzer

FROM debian:bookworm-slim
COPY --from=builder /ocpp-go/mayhem/fuzzOcpp.libfuzzer /
