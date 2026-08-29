# web-stage
FROM node:22-bookworm AS web-stage

WORKDIR /web
COPY cmdb/web/package.json cmdb/web/package-lock.json ./
RUN npm ci
COPY cmdb/web ./
RUN npx vite build

# build-stage
FROM golang:1.25-bookworm AS build-stage

WORKDIR /src
COPY sdk /src/sdk
COPY cmdb /src/cmdb
COPY --from=web-stage /web/dist /src/cmdb/web/dist
WORKDIR /src/cmdb

ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto
RUN go build -o /out/cmdb-agent .

# deploy-stage
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    iputils-ping \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build-stage /out/cmdb-agent ./cmdb-agent

VOLUME /data

EXPOSE 8085

HEALTHCHECK --interval=30s --start-period=15s --timeout=10s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:8085/health || exit 1

ENTRYPOINT ["./cmdb-agent"]
CMD ["--listen", ":8085", "--db-path", "/data/cmdb.db"]
