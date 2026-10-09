# deploy-stage — binary is built by `make build-linux` (same as the main project)
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

COPY bin/backup-agent ./backup-agent

EXPOSE 8087

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8087/health || exit 1

ENTRYPOINT ["./backup-agent"]
CMD ["--listen", ":8087"]
