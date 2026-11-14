# deploy-stage — binary is built by `make build-linux` (embeds web/static).
# pdftoppm/pdftotext (poppler-utils) are required at runtime for PDF page
# rasterization (see rasterize.go); no LibreOffice, so non-PDF documents
# fall back to a text-only diff.
# tesseract-ocr(+rus) — OCR of scanned АОСР for the ИД check (idcheck/);
# optional at runtime, without it pages are classified by the vision model.
# DWG: LibreDWG (dwg2SVG, built from source — not packaged for ubuntu 22.04)
# + librsvg2-bin convert a drawing to PDF, see dwg2pdf.sh / dwg.go.
FROM ubuntu:22.04 AS libredwg-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates curl xz-utils build-essential \
 && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://github.com/LibreDWG/libredwg/releases/download/0.13.3/libredwg-0.13.3.tar.xz | tar -xJ -C /tmp \
 && cd /tmp/libredwg-0.13.3 \
 && ./configure --prefix=/opt/libredwg --disable-bindings --disable-write \
 && make -j"$(nproc)" && make install

FROM ubuntu:22.04 AS deploy-stage

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
    ca-certificates \
    curl \
    librsvg2-bin \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-rus \
 && rm -rf /var/lib/apt/lists/*

COPY --from=libredwg-stage /opt/libredwg /opt/libredwg
ENV PATH="/opt/libredwg/bin:${PATH}" LD_LIBRARY_PATH="/opt/libredwg/lib"
COPY dwg2pdf.sh /usr/local/bin/stroykontrol-dwg2pdf

WORKDIR /app

COPY bin/stroykontrol-web ./stroykontrol-web

EXPOSE 8092

ENTRYPOINT ["./stroykontrol-web"]
CMD ["--listen", ":8092"]
