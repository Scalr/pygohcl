from pybindgen import retval, param, Module
import sys

mod = Module('pygohcl')
mod.add_include('"pygohcl_go.h"')
mod.add_function('Parse', retval('PyObject *', caller_owns_return=False), [param('PyObject *', 'a',  transfer_ownership=False)])
mod.generate(sys.stdout)
