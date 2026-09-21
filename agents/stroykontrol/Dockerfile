# deploy-stage — binary is built by `make build-linux` (embeds web/static).
# pdftoppm/pdftotext (poppler-utils) are required at runtime for PDF page
# rasterization (see rasterize.go); no LibreOffice, so non-PDF documents
# fall back to a text-only diff.
FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    poppler-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY bin/stroykontrol-web ./stroykontrol-web

EXPOSE 8092

ENTRYPOINT ["./stroykontrol-web"]
CMD ["--listen", ":8092"]
