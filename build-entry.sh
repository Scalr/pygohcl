#!/usr/bin/env bash
pythons="cp36-cp36m cp37-cp37m cp38-cp38"

cd /tmp

export PATH="/opt/go/bin:$PATH" HOME=/tmp
for py in $pythons; do
    pybin="/opt/python/${py}/bin"
    "${pybin}/pip" install -r /app/dev-requirements.txt
    "${pybin}/pip" wheel --no-deps --wheel-dir /tmp /dist/*.tar.gz
done
ls *.whl | xargs -n1 --verbose auditwheel repair --wheel-dir /dist

# Install packages and test
for py in $pythons; do
    pybin="/opt/python/${py}/bin"
    "${pybin}/pip" install pygohcl --no-index -f /dist
    "${pybin}/pytest" -p no:cacheprovider /app/tests
done
ls -al /dist
