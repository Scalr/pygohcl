import pytest
import pygohcl


def test_basic():
    assert pygohcl.loads('variable "test" {}') == {"variable": {"test": {}}}


def test_parse_error():
    with pytest.raises(pygohcl.HCLParseError):
        pygohcl.loads('variable "test {}')


def test_internal_error():
    with pytest.raises(pygohcl.HCLInternalError) as exc:
        pygohcl.loads('variable "test" {{test = "{}"}}'.format("a" * 1024 * 1024))

    assert str(exc.value) == "size of HCL file is above maximal size (1048603 > 1048576)"


def test_empty_list():
    out = pygohcl.loads(
        """variable "test" {
    default = []
    }"""
    )
    assert out["variable"]["test"]["default"] == []


def test_null():
    out = pygohcl.loads(
        """variable "test" {
    default = null
    }"""
    )
    assert out["variable"]["test"]["default"] == "null"
