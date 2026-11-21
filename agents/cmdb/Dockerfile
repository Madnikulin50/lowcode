# deploy-stage — binary is built by `make build-linux` (embeds web/dist)
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    iputils-ping \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY bin/cmdb-agent ./cmdb-agent

VOLUME /data

EXPOSE 8085

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8085/health || exit 1

ENTRYPOINT ["./cmdb-agent"]
CMD ["--listen", ":8085", "--db-path", "/data/cmdb.db"]
