package common

// This intentionally bounded Draft 2020-12 subset is the public submission
// contract, not a general schema engine. Admission rejects unknown vocabulary;
// validation never fetches references and never rounds JSON numbers to float64.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var DefaultSubmitSchema = json.RawMessage(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`)

func boundedJSON(raw []byte, byteLimit, nodeLimit, depthLimit int) (any, []byte, int, error) {
	if !utf8.Valid(raw) {
		return nil, nil, 0, fmt.Errorf("JSON must be valid UTF-8")
	}
	if len(raw) > byteLimit {
		return nil, nil, 0, fmt.Errorf("JSON exceeds %d bytes", byteLimit)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		return nil, nil, 0, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, nil, 0, fmt.Errorf("expected one JSON value")
	}
	nodes := 0
	var visit func(any, int) error
	visit = func(v any, depth int) error {
		nodes++
		if nodes > nodeLimit || depth > depthLimit {
			return fmt.Errorf("JSON exceeds node/depth budget")
		}
		switch x := v.(type) {
		case json.Number:
			text := string(x)
			if len(text) > 128 {
				return fmt.Errorf("numeric literal exceeds 128 bytes")
			}
			if i := strings.IndexAny(text, "eE"); i >= 0 {
				e, err := strconv.ParseInt(text[i+1:], 10, 32)
				if err != nil || e < -1000 || e > 1000 {
					return fmt.Errorf("numeric exponent exceeds 1000")
				}
			}
		case map[string]any:
			for _, v := range x {
				if err := visit(v, depth+1); err != nil {
					return err
				}
			}
		case []any:
			for _, v := range x {
				if err := visit(v, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visit(v, 0); err != nil {
		return nil, nil, 0, err
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(canonical) > byteLimit {
		return nil, nil, 0, fmt.Errorf("canonical JSON exceeds byte budget")
	}
	return v, canonical, nodes, nil
}

// CanonicalSubmitSchema owns the schema bytes. Empty means the default contract;
// it stays empty so default result normalization differs from explicit custom schemas.
func CanonicalSubmitSchema(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	v, canonical, _, err := boundedJSON(raw, 65536, 2048, 64)
	if err != nil {
		return nil, err
	}
	root, ok := v.(map[string]any)
	if !ok || root["type"] != "object" {
		return nil, fmt.Errorf("submission schema root must explicitly have type object")
	}
	if err := checkSchema(v, "$"); err != nil {
		return nil, err
	}
	return canonical, nil
}
func CheckSubmitSchema(raw json.RawMessage) error { _, err := CanonicalSubmitSchema(raw); return err }
func rational(v any) (*big.Rat, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return nil, false
	}
	r, ok := new(big.Rat).SetString(string(n))
	return r, ok
}
func integerBound(v any) bool { r, ok := rational(v); return ok && r.IsInt() && r.Sign() >= 0 }
func typeName(t string) bool {
	switch t {
	case "object", "array", "string", "number", "integer", "boolean", "null":
		return true
	}
	return false
}
func checkSchema(v any, path string) error {
	if _, ok := v.(bool); ok {
		return nil
	}
	s, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: schema must be object or boolean", path)
	}
	bad := func(k string) error { return fmt.Errorf("%s: invalid %s", path, k) }
	for k, v := range s {
		switch k {
		case "$schema":
			if v != "https://json-schema.org/draft/2020-12/schema" {
				return bad(k)
			}
		case "title", "description":
			if _, ok := v.(string); !ok {
				return bad(k)
			}
		case "default":
		case "examples":
			if _, ok := v.([]any); !ok {
				return bad(k)
			}
		case "type":
			if name, ok := v.(string); ok {
				if !typeName(name) {
					return bad(k)
				}
			} else {
				names, ok := v.([]any)
				if !ok || len(names) == 0 {
					return bad(k)
				}
				seen := map[string]bool{}
				for _, n := range names {
					name, ok := n.(string)
					if !ok || !typeName(name) || seen[name] {
						return bad(k)
					}
					seen[name] = true
				}
			}
		case "properties":
			props, ok := v.(map[string]any)
			if !ok {
				return bad(k)
			}
			for name, child := range props {
				if err := checkSchema(child, path+".properties."+name); err != nil {
					return err
				}
			}
		case "required":
			names, ok := v.([]any)
			if !ok {
				return bad(k)
			}
			seen := map[string]bool{}
			for _, n := range names {
				name, ok := n.(string)
				if !ok || seen[name] {
					return bad(k)
				}
				seen[name] = true
			}
		case "additionalProperties", "items":
			if err := checkSchema(v, path+"."+k); err != nil {
				return err
			}
		case "minLength", "maxLength", "minItems", "maxItems", "minProperties", "maxProperties":
			if !integerBound(v) {
				return bad(k)
			}
		case "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum":
			if _, ok := rational(v); !ok {
				return bad(k)
			}
		case "multipleOf":
			if n, ok := rational(v); !ok || n.Sign() <= 0 {
				return bad(k)
			}
		case "uniqueItems":
			if _, ok := v.(bool); !ok {
				return bad(k)
			}
		case "pattern":
			p, ok := v.(string)
			if !ok {
				return bad(k)
			}
			if _, err := regexp.Compile(p); err != nil {
				return bad(k)
			}
		case "enum":
			values, ok := v.([]any)
			if !ok || len(values) == 0 {
				return bad(k)
			}
			seen := map[string]bool{}
			for _, v := range values {
				key := valueKey(v)
				if seen[key] {
					return bad(k)
				}
				seen[key] = true
			}
		case "const":
		default:
			return fmt.Errorf("%s: unsupported schema keyword %q", path, k)
		}
	}
	return nil
}

// valueKey represents JSON equality, including exact number equality (1 == 1.0).
// Type tags prevent strings/objects from colliding with numeric representations.
func valueKey(v any) string {
	switch x := v.(type) {
	case json.Number:
		r, _ := rational(x)
		return "n" + r.RatString()
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteByte('o')
		for _, k := range keys {
			b.WriteString(strconv.Quote(k))
			b.WriteString(strconv.Quote(valueKey(x[k])))
		}
		return b.String()
	case []any:
		var b strings.Builder
		b.WriteByte('a')
		for _, v := range x {
			b.WriteString(strconv.Quote(valueKey(v)))
		}
		return b.String()
	default:
		b, _ := json.Marshal(v)
		return "v" + string(b)
	}
}
func matchesType(t string, v any) bool {
	switch t {
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "null":
		return v == nil
	case "number":
		_, ok := rational(v)
		return ok
	case "integer":
		r, ok := rational(v)
		return ok && r.IsInt()
	}
	return false
}
func validateSchema(schema, value any, path string) error {
	if b, ok := schema.(bool); ok {
		if b {
			return nil
		}
		return fmt.Errorf("%s: false schema", path)
	}
	s := schema.(map[string]any)
	fail := func(k string) error { return fmt.Errorf("%s: violates %s", path, k) }
	if typ, ok := s["type"]; ok {
		match := false
		if name, ok := typ.(string); ok {
			match = matchesType(name, value)
		} else {
			for _, name := range typ.([]any) {
				match = match || matchesType(name.(string), value)
			}
		}
		if !match {
			return fail("type")
		}
	}
	if c, ok := s["const"]; ok && valueKey(c) != valueKey(value) {
		return fail("const")
	}
	if values, ok := s["enum"].([]any); ok {
		match := false
		for _, v := range values {
			match = match || valueKey(v) == valueKey(value)
		}
		if !match {
			return fail("enum")
		}
	}
	bounds := func(n int, min, max string) error {
		r := new(big.Rat).SetInt64(int64(n))
		if v, ok := s[min]; ok {
			bound, _ := rational(v)
			if r.Cmp(bound) < 0 {
				return fail(min)
			}
		}
		if v, ok := s[max]; ok {
			bound, _ := rational(v)
			if r.Cmp(bound) > 0 {
				return fail(max)
			}
		}
		return nil
	}
	switch x := value.(type) {
	case map[string]any:
		if err := bounds(len(x), "minProperties", "maxProperties"); err != nil {
			return err
		}
		if required, ok := s["required"].([]any); ok {
			for _, name := range required {
				if _, ok := x[name.(string)]; !ok {
					return fmt.Errorf("%s.%s is required", path, name)
				}
			}
		}
		props, _ := s["properties"].(map[string]any)
		for name, v := range x {
			child, known := props[name]
			if !known {
				child, known = s["additionalProperties"]
			}
			if known {
				if err := validateSchema(child, v, path+"."+name); err != nil {
					return err
				}
			}
		}
	case []any:
		if err := bounds(len(x), "minItems", "maxItems"); err != nil {
			return err
		}
		if unique, _ := s["uniqueItems"].(bool); unique {
			seen := map[string]bool{}
			for _, v := range x {
				k := valueKey(v)
				if seen[k] {
					return fail("uniqueItems")
				}
				seen[k] = true
			}
		}
		if child, ok := s["items"]; ok {
			for i, v := range x {
				if err := validateSchema(child, v, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	case string:
		if err := bounds(utf8.RuneCountInString(x), "minLength", "maxLength"); err != nil {
			return err
		}
		if p, ok := s["pattern"].(string); ok {
			if !regexp.MustCompile(p).MatchString(x) {
				return fail("pattern")
			}
		}
	case json.Number:
		n, _ := rational(x)
		for _, k := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf"} {
			if v, ok := s[k]; ok {
				bound, _ := rational(v)
				cmp := n.Cmp(bound)
				if (k == "minimum" && cmp < 0) || (k == "maximum" && cmp > 0) || (k == "exclusiveMinimum" && cmp <= 0) || (k == "exclusiveMaximum" && cmp >= 0) || (k == "multipleOf" && !new(big.Rat).Quo(n, bound).IsInt()) {
					return fail(k)
				}
			}
		}
	}
	return nil
}

// PrepareSubmission validates and returns independently owned canonical payload.
func PrepareSubmission(schema, payload json.RawMessage) (json.RawMessage, error) {
	canonical, err := CanonicalSubmitSchema(schema)
	if err != nil {
		return nil, err
	}
	isDefault := len(canonical) == 0
	if isDefault {
		canonical = DefaultSubmitSchema
	}
	s, canonical, _, err := boundedJSON(canonical, 65536, 2048, 64)
	if err != nil {
		return nil, err
	}
	value, raw, _, err := boundedJSON(payload, 1048576, 65536, 64)
	if err != nil {
		return nil, err
	}
	if int64(len(raw))*int64(len(canonical)) > 67108864 {
		return nil, fmt.Errorf("submission validation exceeds canonical byte-product budget")
	}
	if err := validateSchema(s, value, "$"); err != nil {
		return nil, err
	}
	if isDefault {
		raw, err = json.Marshal(map[string]any{"result": value.(map[string]any)["result"]})
	}
	return raw, err
}
func ValidateSubmission(schema, payload json.RawMessage) error {
	_, err := PrepareSubmission(schema, payload)
	return err
}
