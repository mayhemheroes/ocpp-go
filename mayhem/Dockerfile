FROM golang:1.18 as builder

COPY . /ocpp-go
WORKDIR /ocpp-go/mayhem

RUN go install github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest
RUN go get github.com/dvyukov/go-fuzz/go-fuzz-dep
RUN apt update && apt install -y clang

RUN go-fuzz-build -libfuzzer -o fuzzOcpp.a && \
    clang -fsanitize=fuzzer fuzzOcpp.a -o fuzzOcpp.libfuzzer

FROM debian:bookworm-slim
COPY --from=builder /ocpp-go/mayhem/fuzzOcpp.libfuzzer /