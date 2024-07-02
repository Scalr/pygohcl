import json
import sysconfig
import typing as tp
from pathlib import Path

from pygohcl._pygohcl import ffi


def load_lib():
    suffix = sysconfig.get_config_var("EXT_SUFFIX")

    libpath = Path(__file__).parent.parent / f"pygohcl{suffix}"
    return ffi.dlopen(str(libpath))


lib = load_lib()


class HCLParseError(Exception):
    pass


class HCLInternalError(Exception):
    pass


class ValidationError(Exception):
    pass


class UnknownFunctionError(ValidationError):
    pass


def loadb(data: bytes) -> tp.Dict:
    s = ffi.new("char[]", data)
    ret = lib.Parse(s)
    if ret.err != ffi.NULL:
        err: bytes = ffi.string(ret.err)
        ffi.gc(ret.err, lib.free)
        err = err.decode("utf8")
        if "invalid HCL:" in err:
            raise HCLParseError(err)
        raise HCLInternalError(err)
    ret_json = ffi.string(ret.json)
    ffi.gc(ret.json, lib.free)
    return json.loads(ret_json)


def loads(data: str) -> tp.Dict:
    return loadb(data.encode("utf8"))


def load(stream: tp.IO) -> tp.Dict:
    data = stream.read()
    return loadb(data)


def attributes_loadb(data: bytes) -> tp.Dict:
    """
    Like :func:`pygohcl.loadb`,
    but expects from the input to contain only top-level attributes.

    Example:
        >>> hcl = '''
        ... key1 = "value"
        ... key2 = false
        ... key3 = [1, 2, 3]
        ... '''
        >>> import pygohcl
        >>> print(pygohcl.attributes_loads(hcl))
        {'key1': 'value', 'key2': False, 'key3': [1, 2, 3]}

    :raises HCLParseError: when the provided input cannot be parsed as valid HCL,
        or it contains other blocks, not only attributes.
    """
    s = ffi.new("char[]", data)
    ret = lib.ParseAttributes(s)
    if ret.err != ffi.NULL:
        err: bytes = ffi.string(ret.err)
        ffi.gc(ret.err, lib.free)
        err = err.decode("utf8")
        raise HCLParseError(err)
    ret_json = ffi.string(ret.json)
    ffi.gc(ret.json, lib.free)
    return json.loads(ret_json)


def attributes_loads(data: str) -> tp.Dict:
    return attributes_loadb(data.encode("utf8"))


def attributes_load(stream: tp.IO) -> tp.Dict:
    data = stream.read()
    return attributes_loadb(data)


def eval_var_condition(
    condition: str, error_message: str, variable_name: str, variable_value: str
) -> None:
    """
    This is specific to Terraform/OpenTofu configuration language
    and is meant to evaluate results of the `validation` block of a variable definition.

    This comes with a limited selection of supported functions.
    Terraform/OpenTofu expand this list with their own set
    of useful functions, which will not pass this validation.
    For that reason a separate `UnknownFunctionError` is raised then,
    so the consumer can decide how to treat this case.

    Example:
        >>> import pygohcl
        >>> pygohcl.eval_var_condition(
        ...     condition="var.count < 3",
        ...     error_message="count must be less than 3, but ${var.count} was given",
        ...     variable_name="count",
        ...     variable_value="5",
        ... )
        Traceback (most recent call last):
            ...
        pygohcl.ValidationError: count must be less than 3, but 5 was given

    :raises ValidationError: when the condition expression has not evaluated to `True`
    :raises UnknownFunctionError: when the condition expression refers to a function
        that is not known to the library
    """
    c = ffi.new("char[]", condition.encode("utf8"))
    e = ffi.new("char[]", error_message.encode("utf8"))
    n = ffi.new("char[]", variable_name.encode("utf8"))
    v = ffi.new("char[]", variable_value.encode("utf8"))
    ret = lib.EvalValidationRule(c, e, n, v)
    if ret != ffi.NULL:
        err: bytes = ffi.string(ret)
        ffi.gc(ret, lib.free)
        err = err.decode("utf8")
        if "Call to unknown function" in err:
            raise UnknownFunctionError(err)
        raise ValidationError(err)
