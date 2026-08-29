# build-stage
FROM golang:1.25-bookworm AS build-stage

WORKDIR /src
COPY sdk /src/sdk
COPY backup /src/backup
WORKDIR /src/backup

ENV CGO_ENABLED=0
RUN go build -o /out/backup-agent .

# deploy-stage
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    restic \
    postgresql-client \
    mariadb-client \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build-stage /out/backup-agent ./backup-agent

EXPOSE 8087

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8087/health || exit 1

ENTRYPOINT ["./backup-agent"]
CMD ["--listen", ":8087"]
