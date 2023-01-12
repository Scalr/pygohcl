#!/usr/bin/env sh

set -e

OS_NAME=$(uname)
case $OS_NAME in
  Darwin*)  OS="darwin" ;;
  Linux*)   OS="linux" ;;
  *)        echo "Unexpected OS: $OS_NAME"
            exit 1
            ;;
esac

# It's defined only on macos runner and ends with an architecture
if [ -z "$ARCHFLAGS" ]; then
  ARCH=$(uname -m)
else
  ARCH=$ARCHFLAGS
fi

case $ARCH in
    *amd64)   ARCH="amd64" ;;
    *x86_64)  ARCH="amd64" ;;
    *arm64)   ARCH="arm64" ;;
    *aarch64) ARCH="arm64" ;;
esac

curl "https://storage.googleapis.com/golang/go1.17.${OS}-${ARCH}.tar.gz" --silent --location | tar -xz
export PATH="$(pwd)/go/bin:$PATH"

echo "OS: $(uname -a)"
echo "GO=go1.17.${OS}-${ARCH}.tar.gz"
echo "ARCH=$ARCH"
echo "PATH=$PATH"

