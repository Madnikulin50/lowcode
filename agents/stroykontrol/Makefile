AGENT := stroykontrol
BIN   := stroykontrol-web

include ../Makefile.inc

.PHONY: all web clean

all: web build

web:
	cd web && npm install && npx vite build

build: web
build-linux: web

clean:
	rm -rf bin/ web/dist/
