import distutils.sysconfig
import json
import typing as tp
from pathlib import Path

from pygohcl._pygohcl import ffi


MAX_SIZE: int = 1024 * 1024  # 1 Mb


def load_lib():
    suffix = distutils.sysconfig.get_config_var("EXT_SUFFIX")

    libpath = Path(__file__).parent.parent / f"pygohcl{suffix}"
    return ffi.dlopen(str(libpath))


lib = load_lib()


class HCLParseError(Exception):
    pass


class HCLInternalError(Exception):
    pass


def loadb(data: bytes) -> tp.Dict:
    if len(data) > MAX_SIZE:
        raise HCLInternalError(f"size of HCL file is above maximal size ({len(data)} > {MAX_SIZE})")

    s = ffi.new("char[]", data)
    ret = lib.Parse(s)
    if ret.err != ffi.NULL:
        err = ffi.string(ret.err)
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
