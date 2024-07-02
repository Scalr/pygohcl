package main

// typedef struct {
// char *json;
// char *err;
// } parseResponse;
import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"strings"
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
		return C.parseResponse{nil, C.CString(diagErrorsToString(diags, "invalid HCL: %s"))}
	}
	hclMap, err := convertFile(hclFile)
	if err != nil {
		return C.parseResponse{nil, C.CString(fmt.Sprintf("cannot convert HCL to Go map representation: %s", err))}
	}
	hclInJson, err := json.Marshal(hclMap)
	if err != nil {
		return C.parseResponse{nil, C.CString(fmt.Sprintf("cannot convert Go map representation to JSON: %s", err))}
	}
	resp = C.parseResponse{C.CString(string(hclInJson)), nil}

	return
}

//export ParseAttributes
func ParseAttributes(a *C.char) (resp C.parseResponse) {
	defer func() {
		if err := recover(); err != nil {
			retValue := fmt.Sprintf("panic HCL: %v", err)
			resp = C.parseResponse{nil, C.CString(retValue)}
		}
	}()

	input := C.GoString(a)
	hclFile, parseDiags := hclsyntax.ParseConfig([]byte(input), "tmp.hcl", hcl.InitialPos)
	if parseDiags.HasErrors() {
		return C.parseResponse{nil, C.CString(diagErrorsToString(parseDiags, "invalid HCL: %s"))}
	}

	var diags hcl.Diagnostics
	hclMap := make(jsonObj)
	c := converter{}

	attrs, attrsDiags := hclFile.Body.JustAttributes()
	diags = diags.Extend(attrsDiags)

	for _, attr := range attrs {
		_, valueDiags := attr.Expr.Value(nil)
		diags = diags.Extend(valueDiags)
		if valueDiags.HasErrors() {
			continue
		}

		value, err := c.convertExpression(attr.Expr.(hclsyntax.Expression))
		if err != nil {
			diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Error processing variable value",
				Detail:   fmt.Sprintf("Cannot convert HCL to Go map representation: %s.", err),
				Subject:  attr.NameRange.Ptr(),
			})
			continue
		}

		hclMap[attr.Name] = value
	}

	hclInJson, err := json.Marshal(hclMap)
	if err != nil {
		diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Error preparing JSON result",
			Detail:   fmt.Sprintf("Cannot convert Go map representation to JSON: %s.", err),
		})
		return C.parseResponse{nil, C.CString(diagErrorsToString(diags, ""))}
	}
	if diags.HasErrors() {
		resp = C.parseResponse{C.CString(string(hclInJson)), C.CString(diagErrorsToString(diags, ""))}
	} else {
		resp = C.parseResponse{C.CString(string(hclInJson)), nil}
	}

	return
}

//export EvalValidationRule
func EvalValidationRule(c *C.char, e *C.char, n *C.char, v *C.char) (resp *C.char) {
	defer func() {
		if err := recover(); err != nil {
			retValue := fmt.Sprintf("panic HCL: %v", err)
			resp = C.CString(retValue)
		}
	}()

	condition := C.GoString(c)
	errorMsg := C.GoString(e)
	varName := C.GoString(n)
	varValue := C.GoString(v)

	// First evaluate variable value to get its cty representation

	varValueCty, diags := expressionValue(varValue, nil)
	if diags.HasErrors() {
		if containsError(diags, "Variables not allowed") {
			// Try again to handle the case when a string value was provided without enclosing quotes
			varValueCty, diags = expressionValue(fmt.Sprintf("%q", varValue), nil)
		}
	}
	if diags.HasErrors() {
		return C.CString(diagErrorsToString(diags, "cannot process variable value: %s"))
	}

	// Now evaluate the condition

	hclCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(map[string]cty.Value{
				varName: varValueCty,
			}),
		},
		Functions: knownFunctions,
	}
	conditionCty, diags := expressionValue(condition, hclCtx)
	if diags.HasErrors() {
		return C.CString(diagErrorsToString(diags, "cannot process condition expression: %s"))
	}

	if conditionCty.IsNull() {
		return C.CString("condition expression result is null")
	}

	conditionCty, err := convert.Convert(conditionCty, cty.Bool)
	if err != nil {
		return C.CString("condition expression result must be bool")
	}

	if conditionCty.True() {
		return nil
	}

	// Finally evaluate the error message expression

	var errorMsgValue = "cannot process error message expression"
	errorMsgCty, diags := expressionValue(errorMsg, hclCtx)
	if diags.HasErrors() {
		errorMsgCty, diags = expressionValue(fmt.Sprintf("%q", errorMsg), hclCtx)
	}
	if !diags.HasErrors() && !errorMsgCty.IsNull() {
		errorMsgCty, err = convert.Convert(errorMsgCty, cty.String)
		if err == nil {
			errorMsgValue = errorMsgCty.AsString()
		}
	}
	return C.CString(errorMsgValue)
}

func diagErrorsToString(diags hcl.Diagnostics, format string) string {
	diagErrs := diags.Errs()
	errors := make([]string, 0, len(diagErrs))
	for _, err := range diagErrs {
		errors = append(errors, err.Error())
	}
	if format == "" {
		return strings.Join(errors, ", ")
	}
	return fmt.Sprintf(format, strings.Join(errors, ", "))
}

func containsError(diags hcl.Diagnostics, e string) bool {
	for _, err := range diags.Errs() {
		if strings.Contains(err.Error(), e) {
			return true
		}
	}
	return false
}

func expressionValue(in string, ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	expr, diags := hclsyntax.ParseExpression([]byte(in), "tmp.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return cty.NilVal, diags
	}

	val, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return cty.NilVal, diags
	}

	return val, diags
}

var knownFunctions = map[string]function.Function{
	"abs":             stdlib.AbsoluteFunc,
	"can":             tryfunc.CanFunc,
	"ceil":            stdlib.CeilFunc,
	"chomp":           stdlib.ChompFunc,
	"coalescelist":    stdlib.CoalesceListFunc,
	"compact":         stdlib.CompactFunc,
	"concat":          stdlib.ConcatFunc,
	"contains":        stdlib.ContainsFunc,
	"csvdecode":       stdlib.CSVDecodeFunc,
	"distinct":        stdlib.DistinctFunc,
	"element":         stdlib.ElementFunc,
	"chunklist":       stdlib.ChunklistFunc,
	"flatten":         stdlib.FlattenFunc,
	"floor":           stdlib.FloorFunc,
	"format":          stdlib.FormatFunc,
	"formatdate":      stdlib.FormatDateFunc,
	"formatlist":      stdlib.FormatListFunc,
	"indent":          stdlib.IndentFunc,
	"join":            stdlib.JoinFunc,
	"jsondecode":      stdlib.JSONDecodeFunc,
	"jsonencode":      stdlib.JSONEncodeFunc,
	"keys":            stdlib.KeysFunc,
	"log":             stdlib.LogFunc,
	"lower":           stdlib.LowerFunc,
	"max":             stdlib.MaxFunc,
	"merge":           stdlib.MergeFunc,
	"min":             stdlib.MinFunc,
	"parseint":        stdlib.ParseIntFunc,
	"pow":             stdlib.PowFunc,
	"range":           stdlib.RangeFunc,
	"regex":           stdlib.RegexFunc,
	"regexall":        stdlib.RegexAllFunc,
	"reverse":         stdlib.ReverseListFunc,
	"setintersection": stdlib.SetIntersectionFunc,
	"setproduct":      stdlib.SetProductFunc,
	"setsubtract":     stdlib.SetSubtractFunc,
	"setunion":        stdlib.SetUnionFunc,
	"signum":          stdlib.SignumFunc,
	"slice":           stdlib.SliceFunc,
	"sort":            stdlib.SortFunc,
	"split":           stdlib.SplitFunc,
	"strrev":          stdlib.ReverseFunc,
	"substr":          stdlib.SubstrFunc,
	"timeadd":         stdlib.TimeAddFunc,
	"title":           stdlib.TitleFunc,
	"trim":            stdlib.TrimFunc,
	"trimprefix":      stdlib.TrimPrefixFunc,
	"trimspace":       stdlib.TrimSpaceFunc,
	"trimsuffix":      stdlib.TrimSuffixFunc,
	"try":             tryfunc.TryFunc,
	"upper":           stdlib.UpperFunc,
	"values":          stdlib.ValuesFunc,
	"zipmap":          stdlib.ZipmapFunc,
}

func main() {}
