#!/usr/bin/env sh

ARCH=$(uname -m)
curl "https://storage.googleapis.com/golang/go1.13.linux-${ARCH}.tar.gz" --silent --location | tar -xz

export PATH="$(pwd)/go/bin:$PATH"
