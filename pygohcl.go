package main

// typedef struct {
// char *json;
// char *err;
// } parseResponse;
import "C"
import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
)

//export Parse
func Parse(a *C.char) (resp C.parseResponse) {
    defer func() {
        if err := recover(); err != nil {
			retValue := fmt.Sprintf("panic HCL: %v", err)
			resp = C.parseResponse{nil, C.CString(retValue)}
        }
    }()

	input := C.GoString(a)
	hclFile, diags := hclparse.NewParser().ParseHCL([]byte(input), "tmp.hcl")
	if diags.HasErrors() {
		errors := make([]string, 0, len(diags))
		for _, diag := range diags {
			errors = append(errors, diag.Error())
		}

		return C.parseResponse{nil, C.CString(fmt.Sprintf("invalid HCL: %s", strings.Join(errors, ", ")))}
	}
	hclMap, err := convertFile(hclFile)
	if err != nil {
		return C.parseResponse{nil, C.CString(fmt.Sprintf("cannot convert HCL to Go map representation: %s", err))}
	}
	hclInJson, err := json.Marshal(hclMap)
	if err != nil {
		return C.parseResponse{nil, C.CString(fmt.Sprintf("cannot Go map representation to JSON: %s", err))}
	}
	resp = C.parseResponse{C.CString(string(hclInJson)), nil}

	return
}

func main() {
}
