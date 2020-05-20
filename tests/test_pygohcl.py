import pygohcl


def test_basic():
    assert pygohcl.loads('variable "test" {}') == {"variable": {"test": {}}}
