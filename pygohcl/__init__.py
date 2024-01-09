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
