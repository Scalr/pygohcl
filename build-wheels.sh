#!/usr/bin/env bash
pythons="cp36-cp36m cp37-cp37m cp38-cp38"
build_dir=${PWD}
export PATH="/opt/go/bin:$PATH" HOME=/tmp

# Build source tar
/opt/python/cp38-cp38/bin/python3 setup.py sdist

# Wheels will be written to /tmp before being auditwheel-repaired
cd /tmp
for py in $pythons; do
    pybin="/opt/python/${py}/bin"
    "${pybin}/pip" install -r "${build_dir}/dev-requirements.txt"
    "${pybin}/pip" wheel --no-deps --wheel-dir /tmp "${build_dir}"/dist/*.tar.gz
done
ls ./*.whl | xargs -n1 --verbose auditwheel repair --wheel-dir "${build_dir}/dist"

# Install packages and test
for py in $pythons; do
    pybin="/opt/python/${py}/bin"
    "${pybin}/pip" install pygohcl --no-index -f "${build_dir}/dist"
    "${pybin}/pytest" -p no:cacheprovider "${build_dir}/tests"
done

# Auditwheel-repaired wheels
ls -al "${build_dir}/dist"
