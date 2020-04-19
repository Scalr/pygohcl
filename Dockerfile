FROM ubuntu:18.04

ENV GOPATH=/go
ENV PATH="/usr/lib/go-1.14/bin:${PATH}"

WORKDIR /app

RUN apt update && \
      apt install -y --no-install-recommends software-properties-common dirmngr apt-transport-https && \
      add-apt-repository ppa:longsleep/golang-backports && \
      apt update

RUN apt install -y --no-install-recommends build-essential gcc pkg-config python3 python3-pip python3-dev golang-1.14

RUN pip3 install pybindgen

CMD make build
