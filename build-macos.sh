#!/usr/bin/env bash

set -e

# Run this script in the environment with the pre-installed go language.
# Before using this script you need to set the target version explicitly in the setup.py as below:
#diff --git a/setup.py b/setup.py
#index 93a7c76..2a80a32 100644
#--- a/setup.py
#+++ b/setup.py
#@@ -11,7 +11,8 @@ os.chdir(os.path.dirname(sys.argv[0]) or ".")
#
# setup(
#     name="pygohcl",
#-    use_scm_version=True,
#+    # use_scm_version=True,
#+    version="1.0.8",
#     description="Python bindings for Hashicorp HCL2 Go library",

mkdir -p ./macos-dist
rm -fr ./macos-dist/*.whl
for VERSION in 3.8.10 3.9.16 3.10.10 3.11.3 3.12.0
do
  # Assumes that all python version are pre-installed.
  # pyenv install -v ${VERSION}

  # Create a python virtual environment and install all dev requirements
  pyenv virtualenv ${VERSION} tmp
  pyenv local tmp
  python -m pip install --upgrade pip
  pip install -r dev-requirements.txt
  pip install wheel

  # Build a wheel
  pip wheel --no-deps --use-pep517 -w dist .

  # Copy the wheel to the artifacts directory
  cp dist/*.whl ./macos-dist

  # cleanup a python virtual environment
  rm -fr build .eggs pygohcl.egg-info dist
  pip freeze | awk '{ print $1 }' | xargs pip uninstall -y
  pyenv virtualenv-delete -f tmp

  # pyenv uninstall -f ${VERSION}
done
