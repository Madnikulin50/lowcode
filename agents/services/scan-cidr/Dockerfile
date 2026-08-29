# build-stage
FROM golang:1.25-bookworm AS build-stage

WORKDIR /src
COPY sdk /src/sdk
COPY services/scan-cidr /src/services/scan-cidr
WORKDIR /src/services/scan-cidr

ENV CGO_ENABLED=0
RUN go build -o /out/scan-cidr ./cmd/scan-cidr

# deploy-stage
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    iputils-ping \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build-stage /out/scan-cidr ./scan-cidr

EXPOSE 8089

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8089/health || exit 1

ENTRYPOINT ["./scan-cidr"]
CMD ["--listen", ":8089"]
