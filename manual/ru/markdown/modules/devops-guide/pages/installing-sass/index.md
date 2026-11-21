# Установка Dart Sass

## Среда разработки

Следуйте инструкциям ниже, чтобы установить двоичные файлы dart-sass в зависимости от вашей операционной системы:

[cols="2s,5a"]
|===
| [#scss-local-linux]#[scss-local-linux,Установка на Linux](#scss-local-linux,Установка на Linux)#
|
[source, bash]
# Using Homebrew.
brew install sass/sass/sass

# Using Snap
sudo snap install dart-sass

| [#scss-local-macOS]#[scss-local-macOS,Установка на MacOS](#scss-local-macOS,Установка на MacOS)#
|
[source, bash]
brew install sass/sass/sass

| [#scss-local-windows]#[scss-local-windows,Установка на Windows](#scss-local-windows,Установка на Windows)#
|
[source, bash]
# Using Chocolatey
choco install sass

# Using Scoop
scoop install sass

|===

Кроме того, вы можете скачать двоичные файлы Dart Sass напрямую со [страницы релизов dart-sass](https://github.com/sass/dart-sass/releases).

!!! important
    При использовании предварительно собранных двоичных файлов убедитесь, что Dart Sass добавлен в PATH вашей системы.


## Среда сборки

В среде сборки, особенно при развёртывании с помощью CI/CD-пайплайнов или Docker, вы можете обеспечить доступность Dart Sass следующими способами:

[cols="2s,5a"]
|===
| [#scss-deployment-github-pages]#[scss-deployment-github-pages,CI/CD-развёртывание на GitHub Pages](#scss-deployment-github-pages,CI/CD-развёртывание на GitHub Pages)#
|

Чтобы установить Dart Sass для CI/CD-развёртывания на GitHub Pages, добавьте следующий шаг в ваш файл рабочего процесса:

[source, bash]
- name: Install Dart Sass
  run: sudo snap install dart-sass

| [#scss-deployment-docker]#[scss-deployment-docker, Docker-развёртывание](#scss-deployment-docker, Docker-развёртывание)#
|
Чтобы установить Dart Sass для Docker, добавьте следующий шаг в ваш Dockerfile:

[source, Dockerfile]
# Replace the image with your deployment Docker image
FROM ubuntu:20.04

RUN apt-get -y update \
 && apt-get -y install \
    ca-certificates \
    curl \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local

# Replace the SASS_VERSION with the version you want to install
ARG SASS_VERSION=1.67.0
ARG SASS*URL="https://github.com/sass/dart-sass/releases/download/${SASS*VERSION}/dart-sass-${SASS_VERSION}-linux-x64.tar.gz"

RUN curl -OL $SASS_URL

# Extract the release (if it's an archive)
RUN tar -xzf dart-sass-${SASS_VERSION}-linux-x64.tar.gz

# Clean up downloaded files (optional)
RUN rm -rf dart-sass-${SASS_VERSION}-linux-x64.tar.gz

ENV PATH=$PATH:/usr/local/dart-sass


|===
