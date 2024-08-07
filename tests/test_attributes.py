import pytest
import pygohcl


def test_basic():
    s = """
    var1 = "value"
    var2 = 2
    var3 = true
    """
    assert pygohcl.attributes_loads(s) == {"var1": "value", "var2": 2, "var3": True}


def test_list():
    s = """
    var1 = ["value1", "value2", "value3"]
    var2 = [1, 2, 3]
    var3 = [true, false]
    """
    assert pygohcl.attributes_loads(s) == {
        "var1": ["value1", "value2", "value3"],
        "var2": [1, 2, 3],
        "var3": [True, False],
    }


def test_empty_list():
    s = """
    var = []
    """
    assert pygohcl.attributes_loads(s) == {"var": []}


def test_non_hcl():
    s = """
    <var = ?>
    """
    with pytest.raises(pygohcl.HCLParseError) as err:
        pygohcl.attributes_loads(s)
    assert "invalid HCL" in str(err.value)


def test_non_attributes():
    """
    When the content is mixed with not expected but valid HCL.
    """
    s = """
    var = "value"
    variable "test" {}
    """
    with pytest.raises(pygohcl.HCLParseError) as err:
        pygohcl.attributes_loads(s)
    assert "Blocks are not allowed" in str(err.value)


def test_variable_in_value():
    s = """
    var1 = "value"
    var2 = value
    """
    with pytest.raises(pygohcl.HCLParseError) as err:
        pygohcl.attributes_loads(s)
    assert "Variables not allowed" in str(err.value)


def test_multiple_errors():
    """
    Make sure the processing doesn't stop at first error and all found issues are reported.
    """
    s = """
    var = value
    variable "test" {}
    """
    with pytest.raises(pygohcl.HCLParseError) as err:
        pygohcl.attributes_loads(s)
    assert "Variables not allowed" in str(err.value)
    assert "Blocks are not allowed" in str(err.value)


def test_heredoc():
    s = """
    var = <<EOT
hey
you
EOT
    """
    assert pygohcl.attributes_loads(s) == {"var": "hey\nyou\n"}


def test_complex_objects():
    s = """
    var = [
        {
            a = 1
            b = 2
        },
        {
            a = "1"
            b = false
        },
        {
            a = [1, 2, 3]
            b = [4, 5, 6]
        },
        {
            c = {
                a = 1
                b = [true, false]
            }
        }
    ]
    """
    assert pygohcl.attributes_loads(s) == {
        "var": [
            {
                "a": 1,
                "b": 2,
            },
            {
                "a": "1",
                "b": False,
            },
            {
                "a": [
                    1,
                    2,
                    3,
                ],
                "b": [
                    4,
                    5,
                    6,
                ],
            },
            {
                "c": {
                    "a": 1,
                    "b": [
                        True,
                        False,
                    ],
                },
            },
        ]
    }
