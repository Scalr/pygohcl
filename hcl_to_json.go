// Adapted from https://github.com/tmccombs/hcl2json
package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	ctyconvert "github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type jsonObj map[string]interface{}

// Convert an hcl File to a json serializable object
// This assumes that the body is a hclsyntax.Body
func convertFile(file *hcl.File, keepInterp bool) (jsonObj, error) {
	c := converter{bytes: file.Bytes, keepInterp: keepInterp}
	body := file.Body.(*hclsyntax.Body)
	return c.convertBody(body)
}

type converter struct {
	bytes      []byte
	keepInterp bool
}

func (c *converter) rangeSource(r hcl.Range) string {
	data := string(c.bytes[r.Start.Byte:r.End.Byte])

	// First process block comments
	data = stripBlockComments(data)

	// Then process inline comments
	lines := stripInlineComments(strings.Split(data, "\n"))

	data = strings.Join(lines, " ")
	data = strings.Join(strings.Fields(data), " ")
	return data
}

func (c *converter) convertBody(body *hclsyntax.Body) (jsonObj, error) {
	var err error
	out := make(jsonObj)
	for key, value := range body.Attributes {
		out[key], err = c.convertExpression(value.Expr)
		if err != nil {
			return nil, err
		}
	}

	for _, block := range body.Blocks {
		err = c.convertBlock(block, out)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (c *converter) convertBlock(block *hclsyntax.Block, out jsonObj) error {
	var key string = block.Type

	value, err := c.convertBody(block.Body)
	if err != nil {
		return err
	}

	for _, label := range block.Labels {
		if inner, exists := out[key]; exists {
			var ok bool
			out, ok = inner.(jsonObj)
			if !ok {
				// TODO: better diagnostics
				return fmt.Errorf("unable to convert Block to JSON: %v.%v", block.Type, strings.Join(block.Labels, "."))
			}
		} else {
			obj := make(jsonObj)
			out[key] = obj
			out = obj
		}
		key = label
	}

	if current, exists := out[key]; exists {
		if list, ok := current.([]interface{}); ok {
			out[key] = append(list, value)
		} else {
			out[key] = []interface{}{current, value}
		}
	} else {
		out[key] = value
	}

	return nil
}

func (c *converter) convertExpression(expr hclsyntax.Expression) (interface{}, error) {
	// assume it is hcl syntax (because, um, it is)
	switch value := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyjson.SimpleJSONValue{Value: value.Val}, nil
	case *hclsyntax.UnaryOpExpr:
		return c.convertUnary(value)
	case *hclsyntax.TemplateExpr:
		return c.convertTemplate(value)
	case *hclsyntax.TemplateWrapExpr:
		return c.convertExpression(value.Wrapped)
	case *hclsyntax.TupleConsExpr:
		list := make([]interface{}, 0)
		for _, ex := range value.Exprs {
			elem, err := c.convertExpression(ex)
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		return list, nil
	case *hclsyntax.ObjectConsExpr:
		m := make(jsonObj)
		for _, item := range value.Items {
			key, err := c.convertKey(item.KeyExpr)
			if err != nil {
				return nil, err
			}
			m[key], err = c.convertExpression(item.ValueExpr)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	default:
		return c.wrapExpr(expr), nil
	}
}

func (c *converter) convertTemplate(t *hclsyntax.TemplateExpr) (string, error) {
	if t.IsStringLiteral() {
		// safe because the value is just the string
		v, err := t.Value(nil)
		if err != nil {
			return "", err
		}
		return v.AsString(), nil
	}
	var builder strings.Builder
	for _, part := range t.Parts {
		s, err := c.convertStringPart(part)
		if err != nil {
			return "", err
		}
		builder.WriteString(s)
	}
	return builder.String(), nil
}

func (c *converter) convertStringPart(expr hclsyntax.Expression) (string, error) {
	switch v := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		s, err := ctyconvert.Convert(v.Val, cty.String)
		if err != nil {
			return "", err
		}
		return s.AsString(), nil
	case *hclsyntax.TemplateExpr:
		return c.convertTemplate(v)
	case *hclsyntax.TemplateWrapExpr:
		return c.convertStringPart(v.Wrapped)
	case *hclsyntax.ConditionalExpr:
		return c.convertTemplateConditional(v)
	case *hclsyntax.TemplateJoinExpr:
		return c.convertTemplateFor(v.Tuple.(*hclsyntax.ForExpr))
	case *hclsyntax.ScopeTraversalExpr:
		return c.wrapTraversal(expr), nil

	default:
		// treating as an embedded expression
		return c.wrapExpr(expr), nil
	}
}

func (c *converter) convertKey(keyExpr hclsyntax.Expression) (string, error) {
	// a key should never have dynamic input
	if k, isKeyExpr := keyExpr.(*hclsyntax.ObjectConsKeyExpr); isKeyExpr {
		keyExpr = k.Wrapped
		if _, isTraversal := keyExpr.(*hclsyntax.ScopeTraversalExpr); isTraversal {
			return c.rangeSource(keyExpr.Range()), nil
		}
	}
	return c.convertStringPart(keyExpr)
}

func (c *converter) convertTemplateConditional(expr *hclsyntax.ConditionalExpr) (string, error) {
	var builder strings.Builder
	builder.WriteString("%{if ")
	builder.WriteString(c.rangeSource(expr.Condition.Range()))
	builder.WriteString("}")
	trueResult, err := c.convertStringPart(expr.TrueResult)
	if err != nil {
		return "", nil
	}
	builder.WriteString(trueResult)
	falseResult, _ := c.convertStringPart(expr.FalseResult)
	if len(falseResult) > 0 {
		builder.WriteString("%{else}")
		builder.WriteString(falseResult)
	}
	builder.WriteString("%{endif}")

	return builder.String(), nil
}

func (c *converter) convertTemplateFor(expr *hclsyntax.ForExpr) (string, error) {
	var builder strings.Builder
	builder.WriteString("%{for ")
	if len(expr.KeyVar) > 0 {
		builder.WriteString(expr.KeyVar)
		builder.WriteString(", ")
	}
	builder.WriteString(expr.ValVar)
	builder.WriteString(" in ")
	builder.WriteString(c.rangeSource(expr.CollExpr.Range()))
	builder.WriteString("}")
	templ, err := c.convertStringPart(expr.ValExpr)
	if err != nil {
		return "", err
	}
	builder.WriteString(templ)
	builder.WriteString("%{endfor}")

	return builder.String(), nil
}

func (c *converter) wrapExpr(expr hclsyntax.Expression) string {
	return c.rangeSource(expr.Range())
}

func (c *converter) wrapTraversal(expr hclsyntax.Expression) string {
	res := c.wrapExpr(expr)
	if c.keepInterp {
		res = "${" + res + "}"
	}
	return res
}

func (c *converter) convertUnary(v *hclsyntax.UnaryOpExpr) (interface{}, error) {
	_, isLiteral := v.Val.(*hclsyntax.LiteralValueExpr)
	if !isLiteral {
		return c.wrapExpr(v), nil
	}
	val, err := v.Value(nil)
	if err != nil {
		return nil, err
	}
	return ctyjson.SimpleJSONValue{Value: val}, nil
}

// stripInlineComments removes single-line comments that start with # or // from each line.
func stripInlineComments(lines []string) []string {
	stripped := make([]string, len(lines))
	for i, line := range lines {
		// Track if we're inside a string literal
		inString := false
		// Character used to open the string literal
		inStringChar := byte(0)
		var strippedLine strings.Builder

		for j := 0; j < len(line); j++ {
			char := line[j]

			// Handle string literals
			if char == '"' || char == '\'' {
				if !inString {
					inString = true
					inStringChar = char
				} else if inStringChar == char {
					inString = false
				}
			}

			// Only process comments if we're not inside a string literal
			if !inString {
				if char == '#' || (char == '/' && j+1 < len(line) && line[j+1] == '/') {
					// Found a comment, stop processing this line
					break
				}
			}

			strippedLine.WriteByte(char)
		}

		stripped[i] = strippedLine.String()
	}
	return stripped
}

// stripBlockComments removes block comments that start with /* and end with */ from a given string.
func stripBlockComments(text string) string {
	var stripped strings.Builder
	// Track if we're inside a string literal
	inString := false
	// Character used to open the string literal
	inStringChar := byte(0)
	// Track if we're inside a block comment
	inBlockComment := false

	for i := 0; i < len(text); i++ {
		char := text[i]

		// Handle string literals
		if char == '"' || char == '\'' {
			if !inString {
				inString = true
				inStringChar = char
			} else if inStringChar == char {
				inString = false
			}
		}

		// Only process comments if we're not inside a string literal
		if !inString {
			if !inBlockComment && char == '/' && i+1 < len(text) && text[i+1] == '*' {
				// Found the start of a block comment
				inBlockComment = true
				i++ // Skip the '*'
				continue
			}

			if inBlockComment && char == '*' && i+1 < len(text) && text[i+1] == '/' {
				// Found the end of a block comment
				inBlockComment = false
				i++ // Skip the '/'
				continue
			}
		}

		if !inBlockComment {
			stripped.WriteByte(char)
		}
	}

	return stripped.String()
}
