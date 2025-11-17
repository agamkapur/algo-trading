FROM alpine:latest AS builder

RUN apk add --no-cache \
    g++ \
    cmake \
    make \
    boost-dev \
    boost-json \
    openssl-dev \
    linux-headers

WORKDIR /build
COPY . .

RUN cmake . && make

FROM alpine:latest

RUN apk add --no-cache \
    libstdc++ \
    boost-system \
    boost-json \
    openssl

WORKDIR /app
COPY --from=builder /build/kucoin_connector .

CMD ["./kucoin_connector"]
