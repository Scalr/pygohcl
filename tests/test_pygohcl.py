import pytest
import pygohcl


def test_basic():
    assert pygohcl.loads('variable "test" {}') == {"variable": {"test": {}}}


def test_parse_error():
    with pytest.raises(pygohcl.HCLParseError):
        pygohcl.loads('variable "test {}')


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
    assert out["variable"]["test"]["default"] is None


def test_numbers():
    out = pygohcl.loads(
        """locals {
            a = 0.19
            b = 1 + 9
            c = -0.82
            x = -10
            y = -x
            z = -(1 + 4)
        }"""
    )
    assert out["locals"]["a"] == 0.19
    assert out["locals"]["b"] == "1+9"
    assert out["locals"]["c"] == -0.82
    assert out["locals"]["x"] == -10
    assert out["locals"]["y"] == "-x"
    assert out["locals"]["z"] == "-(1+4)"


def test_value_is_null():
    with pytest.raises(pygohcl.HCLInternalError):
        pygohcl.loads(
            """resource "datadog_synthetics_test" "status_check_api" {
                message = <<-EOT
                    ${local.is_production_env ? "prod" : null}
                EOT
            }"""
        )


def test_namespaced_functions():
    assert pygohcl.loads(
    """locals {
        timestamp = provider::time::rfc3339_parse(plantimestamp())
    }""") == {"locals": {"timestamp": "provider::time::rfc3339_parse(plantimestamp())"}}
