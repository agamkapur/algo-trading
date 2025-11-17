FROM alpine:latest AS builder

RUN apk add --no-cache \
    g++ \
    cmake \
    make \
    boost-dev \
    openssl-dev \
    linux-headers

WORKDIR /build
COPY . .

RUN cmake . && make

FROM alpine:latest

RUN apk add --no-cache \
    libstdc++ \
    boost-system \
    openssl

WORKDIR /app
COPY --from=builder /build/bitget_connector .

CMD ["./bitget_connector"]
