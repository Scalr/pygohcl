[![pygohcl python package](https://github.com/Scalr/pygohcl/actions/workflows/default.yml/badge.svg)](https://github.com/Scalr/pygohcl/actions/workflows/default.yml)

# pygohcl
Python wrapper for [hashicorp/hcl](https://github.com/hashicorp/hcl) (v2).

## Requirements
The following versions are supported - 3.6, 3.7, 3.8, 3.9, 3.10, 3.11, 3.12.

## Setup
```sh
pip install pygohcl
```

## Usage
```py
>>> import pygohcl
>>> pygohcl.loads("""variable "docker_ports" {
...   type = list(object({
...     internal = number
...     external = number
...     protocol = string
...   }))
...   default = [
...     {
...       internal = 8300
...       external = 8300
...       protocol = "tcp"
...     }
...   ]
... }""")
{'variable': {'docker_ports': {'default': [{'external': 8300, 'internal': 8300, 'protocol': 'tcp'}], 'type': 'list(object({internal=numberexternal=numberprotocol=string}))'}}}
```

## Building locally
You can use the following commands to build a wheel for your platform:
```sh
pip install wheel
python setup.py bdist_wheel
```

The wheel will be available in `./dist/`.
