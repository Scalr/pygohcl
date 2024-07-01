import pytest
import pygohcl


def test_basic_success():
    c = "var.count < 3"
    e = "must be less than 3"
    n = "count"
    v = "1"

    pygohcl.eval_var_condition(c, e, n, v)


def test_basic_fail():
    c = "var.count < 3"
    e = "must be less than 3"
    n = "count"
    v = "5"

    with pytest.raises(pygohcl.ValidationError) as err:
        pygohcl.eval_var_condition(c, e, n, v)

    assert str(err.value) == "must be less than 3"


def test_error_message_eval():
    c = "var.count < 3"
    e = "must be less than 3, ${var.count} was given"
    n = "count"
    v = "5"

    with pytest.raises(pygohcl.ValidationError) as err:
        pygohcl.eval_var_condition(c, e, n, v)

    assert str(err.value) == "must be less than 3, 5 was given"


def test_error_message_failed_eval():
    c = "var.count < 3"
    e = "${var.missing}"
    n = "count"
    v = "5"

    with pytest.raises(pygohcl.ValidationError) as err:
        pygohcl.eval_var_condition(c, e, n, v)

    assert str(err.value) == "cannot process error message expression"


def test_function():
    c = 'contains(["Windows", "Linux"], var.os_type)'
    e = "The os_type must be either 'Windows' or 'Linux'."
    n = "os_type"
    v = "Linux"

    pygohcl.eval_var_condition(c, e, n, v)


def test_unknown_function():
    c = "sin(var.angle) > 0.7"
    e = "invalid angle"
    n = "angle"
    v = "45"

    with pytest.raises(pygohcl.UnknownFunctionError) as err:
        pygohcl.eval_var_condition(c, e, n, v)
