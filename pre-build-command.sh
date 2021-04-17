#!/usr/bin/env sh

curl https://storage.googleapis.com/golang/go1.13.linux-amd64.tar.gz --silent --location | tar -xz

export PATH="$(pwd)/go/bin:$PATH"
