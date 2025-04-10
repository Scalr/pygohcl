import pytest
import pygohcl


@pytest.mark.parametrize(
    "hcl,expected",
    [
        (  # Simple inline comments
            """variable "test" {
                validation {
                    condition = alltrue(
                        # This is a comment on it's own line
                        contains(["a", "b"], var) // This is a trailing comment
                    )
                }
            }""",
            {"variable": {"test": {"validation": {"condition": "alltrue( contains([\"a\", \"b\"], var) )"}}}},
        ),
        (  # Inline comment inside of a string literal
            """variable "test" {
                validation {
                    condition = contains(["a", "b #not a comment"], var) # This is a comment
                }
            }""",
            {"variable": {"test": {"validation": {"condition": "contains([\"a\", \"b #not a comment\"], var)"}}}},
        ),
        (  # Block comment
            """variable "test" {
                validation {
                    condition = alltrue(
                        /* This is a
                            block comment */
                        contains(["a", "b"], var)
                    )
                }
            }""",
            {"variable": {"test": {"validation": {"condition": "alltrue( contains([\"a\", \"b\"], var) )"}}}},
        ),
        (  # Block comment inside of a string literal
            """variable "test" {
                validation {
                    condition = contains(["a", "b /*not a comment*/"], var) /* This is a comment */
                }
            }""",
            {"variable": {"test": {"validation": {"condition": "contains([\"a\", \"b /*not a comment*/\"], var)"}}}},
        ),
        (  # Mixed comments, nested string literals
            """variable "test" {
                validation {
                    condition = alltrue([ # Who
                        for val in var.values: // would
                            /* comment
                               like */
                            !contains(["'#a'", "//b", "\\"c\\""], val.a) || # this?!
                            contains(val.b, "/*x*/")
                    ])
                }
            }""",
            {
                "variable": {
                    "test": {
                        "validation": {
                            "condition": "alltrue([ for val in var.values: !contains([\"'#a'\", \"//b\", \"\\\"c\\\"\"], val.a) || contains(val.b, \"/*x*/\") ])",
                        }
                    }
                }
            },
        ),
    ],
)
def test_expression_comments(hcl, expected):
    assert pygohcl.loads(hcl, keep_interpolations=True) == expected
