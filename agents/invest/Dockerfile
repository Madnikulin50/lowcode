# build-stage
FROM golang:1.25-bookworm AS build-stage

WORKDIR /src
COPY services/calc-evm /src/services/calc-evm
COPY sdk /src/sdk
COPY invest /src/invest
WORKDIR /src/invest

ENV CGO_ENABLED=0
RUN go build -o /out/invest-engine .

# deploy-stage
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build-stage /out/invest-engine ./invest-engine

EXPOSE 8086

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8086/api/health || exit 1

ENTRYPOINT ["./invest-engine"]
CMD ["--listen", ":8086"]
