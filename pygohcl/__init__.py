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


def tfvars_loadb(data: bytes) -> tp.Dict:
    s = ffi.new("char[]", data)
    ret = lib.ParseTfVars(s)
    if ret.err != ffi.NULL:
        err: bytes = ffi.string(ret.err)
        ffi.gc(ret.err, lib.free)
        err = err.decode("utf8")
        raise HCLParseError(err)
    ret_json = ffi.string(ret.json)
    ffi.gc(ret.json, lib.free)
    return json.loads(ret_json)


def tfvars_loads(data: str) -> tp.Dict:
    return tfvars_loadb(data.encode("utf8"))


def tfvars_load(stream: tp.IO) -> tp.Dict:
    data = stream.read()
    return tfvars_loadb(data)


def eval_var_condition(
    condition: str, error_message: str, variable_name: str, variable_value: str
) -> None:
    c = ffi.new("char[]", condition.encode("utf8"))
    e = ffi.new("char[]", error_message.encode("utf8"))
    n = ffi.new("char[]", variable_name.encode("utf8"))
    v = ffi.new("char[]", variable_value.encode("utf8"))
    ret = lib.EvalValidationRule(c,e,n,v)
    if ret != ffi.NULL:
        err: bytes = ffi.string(ret)
        ffi.gc(ret, lib.free)
        err = err.decode("utf8")
        if "Call to unknown function" in err:
            raise UnknownFunctionError(err)
        raise ValidationError(err)
