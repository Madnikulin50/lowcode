# build-stage
FROM golang:1.25-bookworm AS build-stage

WORKDIR /src
COPY sdk /src/sdk
COPY services/calc-evm /src/services/calc-evm
WORKDIR /src/services/calc-evm

ENV CGO_ENABLED=0
RUN go build -o /out/calc-evm ./cmd/calc-evm

# deploy-stage
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build-stage /out/calc-evm ./calc-evm

EXPOSE 8088

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8088/health || exit 1

ENTRYPOINT ["./calc-evm"]
CMD ["--listen", ":8088"]
