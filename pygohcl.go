package main

// #cgo pkg-config: python-3.6
// #define Py_LIMITED_API
// #include <Python.h>
import "C"
import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
)

func GoString_FromPyString(t *C.PyObject) (string, bool) {
	unicodePystr := C.PyUnicode_FromObject(t)
	if unicodePystr == nil {
		return "", false
	}
	bytePystr := C.PyUnicode_AsASCIIString(unicodePystr)
	if bytePystr == nil {
		return "", false
	}
	typePystr := C.PyBytes_AsString(bytePystr)
	if typePystr == nil {
		return "", false
	}
	return C.GoString(typePystr), true
}

//export Parse
func Parse(a *C.PyObject) *C.PyObject {
	input, _ := GoString_FromPyString(a)
	_, diags := hclparse.NewParser().ParseHCL([]byte(input), "tmp.hcl")
	if len(diags) == 0 {
		return C.PyUnicode_FromString(C.CString("valid HCL"))
	} else {
		errors := make([]string, 0, len(diags))
		for _, diag := range diags {
			errors = append(errors, diag.Error())
		}

		return C.PyUnicode_FromString(C.CString(fmt.Sprintf("invalid HCL: %s", strings.Join(errors, ", "))))
	}
}

func main() {
}
