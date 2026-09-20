# build-stage
FROM alpine:3 as build-stage


ARG SASS_VERSION=1.99.0
RUN apk update && apk add --no-cache file


ARG SASS_URL=https://github.com/sass/dart-sass/releases/download/1.69.5/dart-sass-${SASS_VERSION}-linux-x64.tar.gz

WORKDIR /tmp
COPY ./dart-sass-${SASS_VERSION}-linux-x64.tar.gz ./
RUN ls ./

RUN tar -xzf dart-sass-${SASS_VERSION}-linux-x64.tar.gz

ARG VERSION=2026.09.20

ARG SERVER_VERSION=${VERSION}
ARG WEBAPP_VERSION=${VERSION}

RUN mkdir /pnp/
ADD ./server/dist/lowcode-server-${VERSION}-linux-amd64 /pnp/lowcode-server--linux-amd64

WORKDIR /pnp
RUN mkdir /pnp/provision
ADD ./server/provision /pnp/provision/

RUN rm -rf /pnp/webapp

RUN mkdir /pnp/webapp
RUN mkdir /pnp/webapp/admin
RUN mkdir /pnp/webapp/compose
RUN mkdir /pnp/webapp/discovery
RUN mkdir /pnp/webapp/privacy
RUN mkdir /pnp/webapp/reporter
RUN mkdir /pnp/webapp/workflow

ADD ./client3/web/one/dist /pnp/webapp/
ADD ./client3/web/admin/dist /pnp/webapp/admin
ADD ./client3/web/compose/dist /pnp/webapp/compose
ADD ./client3/web/discovery/dist /pnp/webapp/discovery
ADD ./client3/web/privacy/dist /pnp/webapp/privacy
ADD ./client3/web/reporter/dist /pnp/webapp/reporter
ADD ./client3/web/workflow/dist /pnp/webapp/workflow

RUN test -s /pnp/webapp/compose/index.html || (echo "ERROR: compose webapp index.html is missing or empty" && exit 1)


# libredwg-stage: builds dwg2dxf (DWG -> DXF conversion) from source, since
# no distro package exists for Ubuntu 22.04. Pinned to the 0.14 release tag.
FROM ubuntu:22.04 as libredwg-stage

RUN apt-get -y update && apt-get -y install \
    build-essential autoconf automake libtool pkg-config \
    git ca-certificates texinfo python3 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /src
RUN git clone --branch 0.14 --depth 1 https://github.com/LibreDWG/libredwg.git .
RUN ./autogen.sh
RUN ./configure --disable-bindings --disable-static --without-libintl-prefix --without-libiconv-prefix
RUN make -j"$(nproc)"
RUN make DESTDIR=/out install


# deploy-stage
FROM ubuntu:22.04 as deploy-stage

RUN apt-get -y update \
 && apt-get -y install \
    ca-certificates \
    curl \
 && rm -rf /var/lib/apt/lists/*

COPY --from=libredwg-stage /out/usr/local/bin/dwg2dxf /usr/local/bin/dwg2dxf
COPY --from=libredwg-stage /out/usr/local/lib/libredwg.so.0.0.14 /usr/local/lib/libredwg.so.0
RUN ldconfig

ENV STORAGE_PATH="/data"
ENV CORREDOR_ADDR="corredor:80"
ENV HTTP_ADDR="0.0.0.0:80"
ENV HTTP_WEBAPP_ENABLED="true"
ENV HTTP_WEBAPP_BASE_DIR="/pnp/webapp"
ENV PATH="/opt/dart-sass:/pnp/bin:${PATH}"

WORKDIR /pnp

VOLUME /data

COPY --from=build-stage /pnp ./
COPY --from=build-stage /tmp/dart-sass /opt/dart-sass

HEALTHCHECK --interval=30s --start-period=1m --timeout=30s --retries=3 \
    CMD curl --silent --fail --fail-early http://127.0.0.1:80/healthcheck || exit 1

EXPOSE 80

RUN ls .

ENTRYPOINT ["./lowcode-server--linux-amd64"]

CMD ["serve-api"]
