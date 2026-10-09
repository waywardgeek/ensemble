package mcp

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// A compiled schema is immutable and retains its creating service for syntax
// and diagnostics. Nodes are physical schema locations, not expanded copies.
type schema struct {
	parent common.MCPService
	root   any
	nodes  map[string]*schemaNode
	entry  *schemaNode
}
type schemaNode struct {
	value    any
	path     string
	children []*schemaNode
	ref      *schemaNode
	height   int
}

func compile(s common.MCPService, raw []byte, input bool) (*schema, error) {
	v, err := s.Ensemble().JSON().Parse(raw, common.JSONBounds{Bytes: common.MCPMessageLimit, Depth: 64, Nodes: 100000, Collection: 100000, Scalars: true})
	if err != nil {
		return nil, failure("mcp_protocol", "Invalid schema JSON.")
	}
	canonical, err := s.Ensemble().JSON().Encode(v, true, 256<<10)
	if err != nil || len(canonical) > 256<<10 {
		return nil, failure("mcp_limit", "Schema byte limit.")
	}
	if input {
		m, ok := v.(map[string]any)
		if !ok || m["type"] != "object" {
			return nil, failure("mcp_protocol", "Input schema must explicitly declare object.")
		}
	}
	sc := &schema{parent: s, root: v, nodes: map[string]*schemaNode{}}
	sc.entry, err = sc.node(v, "")
	if err != nil {
		return nil, err
	}
	// Resolve all reachable physical targets before graph cycle/depth analysis.
	for {
		before := len(sc.nodes)
		for _, n := range sc.nodes {
			if m, ok := n.value.(map[string]any); ok {
				if r, ok := m["$ref"].(string); ok {
					target, path, err := sc.pointer(r)
					if err != nil {
						return nil, err
					}
					n.ref, err = sc.node(target, path)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if before == len(sc.nodes) {
			break
		}
	}
	colors := map[*schemaNode]int{}
	for _, n := range sc.nodes {
		if _, err = sc.depth(n, colors); err != nil {
			return nil, err
		}
	}
	return sc, nil
}
func (sc *schema) bad() error { return failure("mcp_protocol", "Unsupported or malformed schema.") }
func pointerPart(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}
func (sc *schema) node(v any, path string) (*schemaNode, error) {
	if n := sc.nodes[path]; n != nil {
		return n, nil
	}
	if len(sc.nodes) >= 4096 {
		return nil, failure("mcp_limit", "Schema node limit.")
	}
	n := &schemaNode{value: v, path: path}
	sc.nodes[path] = n
	if _, ok := v.(bool); ok {
		return n, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, sc.bad()
	}
	child := func(v any, path string) error {
		x, e := sc.node(v, path)
		if e == nil {
			n.children = append(n.children, x)
		}
		return e
	}
	for k, v := range m {
		p := path + "/" + pointerPart(k)
		switch k {
		case "$schema":
			if v != "https://json-schema.org/draft/2020-12/schema" {
				return nil, sc.bad()
			}
		case "$ref":
			r, ok := v.(string)
			if !ok || r != "#" && !strings.HasPrefix(r, "#/") {
				return nil, sc.bad()
			}
		case "$defs", "properties":
			o, ok := v.(map[string]any)
			if !ok {
				return nil, sc.bad()
			}
			for key, x := range o {
				if e := child(x, p+"/"+pointerPart(key)); e != nil {
					return nil, e
				}
			}
		case "items", "additionalProperties", "not":
			if e := child(v, p); e != nil {
				return nil, e
			}
		case "allOf", "anyOf", "oneOf":
			a, ok := v.([]any)
			if !ok || len(a) == 0 {
				return nil, sc.bad()
			}
			for i, x := range a {
				if e := child(x, p+"/"+strconv.Itoa(i)); e != nil {
					return nil, e
				}
			}
		case "type":
			types := []any{v}
			if a, ok := v.([]any); ok {
				types = a
			}
			if len(types) == 0 {
				return nil, sc.bad()
			}
			seen := map[string]bool{}
			for _, x := range types {
				t, ok := x.(string)
				if !ok || seen[t] {
					return nil, sc.bad()
				}
				switch t {
				case "object", "array", "string", "number", "integer", "boolean", "null":
				default:
					return nil, sc.bad()
				}
				seen[t] = true
			}
		case "required":
			a, ok := v.([]any)
			if !ok {
				return nil, sc.bad()
			}
			seen := map[string]bool{}
			for _, x := range a {
				t, ok := x.(string)
				if !ok || seen[t] {
					return nil, sc.bad()
				}
				seen[t] = true
			}
		case "minItems", "maxItems", "minLength", "maxLength", "minProperties", "maxProperties":
			x, ok := v.(json.Number)
			if !ok || !sc.parent.Ensemble().JSON().Integral(x) || sc.parent.Ensemble().JSON().Compare(x, "0") < 0 {
				return nil, sc.bad()
			}
		case "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum":
			if _, ok := v.(json.Number); !ok {
				return nil, sc.bad()
			}
		case "enum":
			a, ok := v.([]any)
			if !ok || len(a) == 0 {
				return nil, sc.bad()
			}
			seen := map[string]bool{}
			for _, x := range a {
				b, e := sc.parent.Ensemble().JSON().Encode(x, true, 256<<10)
				if e != nil || seen[string(b)] {
					return nil, sc.bad()
				}
				seen[string(b)] = true
			}
		case "title", "description":
			if _, ok := v.(string); !ok {
				return nil, sc.bad()
			}
		case "examples":
			if _, ok := v.([]any); !ok {
				return nil, sc.bad()
			}
		case "const", "default":
		default:
			return nil, sc.bad()
		}
	}
	return n, nil
}
func (sc *schema) pointer(r string) (any, string, error) {
	if r == "#" {
		return sc.root, "", nil
	}
	if !strings.HasPrefix(r, "#/") {
		return nil, "", sc.bad()
	}
	v := sc.root
	path := ""
	for _, escaped := range strings.Split(r[2:], "/") {
		var b strings.Builder
		for i := 0; i < len(escaped); i++ {
			if escaped[i] == '~' {
				i++
				if i >= len(escaped) || escaped[i] != '0' && escaped[i] != '1' {
					return nil, "", sc.bad()
				}
				if escaped[i] == '0' {
					b.WriteByte('~')
				} else {
					b.WriteByte('/')
				}
			} else {
				b.WriteByte(escaped[i])
			}
		}
		key := b.String()
		path += "/" + pointerPart(key)
		switch x := v.(type) {
		case map[string]any:
			var ok bool
			v, ok = x[key]
			if !ok {
				return nil, "", sc.bad()
			}
		case []any:
			i, e := strconv.Atoi(key)
			if e != nil || i < 0 || i >= len(x) || strconv.Itoa(i) != key {
				return nil, "", sc.bad()
			}
			v = x[i]
		default:
			return nil, "", sc.bad()
		}
	}
	return v, path, nil
}
func (sc *schema) depth(n *schemaNode, colors map[*schemaNode]int) (int, error) {
	if colors[n] == 1 {
		return 0, sc.bad()
	}
	if colors[n] == 2 {
		return n.height, nil
	}
	colors[n] = 1
	h := 1
	edges := n.children
	if n.ref != nil {
		edges = append(append([]*schemaNode{}, edges...), n.ref)
	}
	for _, c := range edges {
		d, e := sc.depth(c, colors)
		if e != nil {
			return 0, e
		}
		if d+1 > h {
			h = d + 1
		}
		if h > 64 {
			return 0, failure("mcp_limit", "Schema expansion depth limit.")
		}
	}
	colors[n] = 2
	n.height = h
	return h, nil
}
func (sc *schema) validate(v any) error {
	steps := 0
	ok, err := sc.check(sc.entry, v, &steps)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("arguments/result do not satisfy schema")
	}
	return nil
}
func (sc *schema) check(n *schemaNode, v any, steps *int) (bool, error) {
	*steps++
	if *steps > 100000 {
		return false, failure("mcp_limit", "Schema validation work limit.")
	}
	if b, ok := n.value.(bool); ok {
		return b, nil
	}
	m := n.value.(map[string]any)
	test := func(child *schemaNode, v any) (bool, error) { return sc.check(child, v, steps) }
	if n.ref != nil {
		ok, e := test(n.ref, v)
		if e != nil || !ok {
			return ok, e
		}
	}
	if t, ok := m["type"]; ok {
		ts := []any{t}
		if a, ok := t.([]any); ok {
			ts = a
		}
		match := false
		for _, x := range ts {
			if sc.matches(x.(string), v) {
				match = true
			}
		}
		if !match {
			return false, nil
		}
	}
	eq := func(a, b any) bool {
		x, e := sc.parent.Ensemble().JSON().Encode(a, true, common.MCPMessageLimit)
		if e != nil {
			return false
		}
		y, e := sc.parent.Ensemble().JSON().Encode(b, true, common.MCPMessageLimit)
		return e == nil && string(x) == string(y)
	}
	if c, ok := m["const"]; ok && !eq(c, v) {
		return false, nil
	}
	if a, ok := m["enum"].([]any); ok {
		found := false
		for _, x := range a {
			if eq(x, v) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		if a, ok := m[key].([]any); ok {
			count := 0
			for i := range a {
				ok, e := test(sc.nodes[n.path+"/"+key+"/"+strconv.Itoa(i)], v)
				if e != nil {
					return false, e
				}
				if ok {
					count++
				}
			}
			if key == "allOf" && count != len(a) || key == "anyOf" && count == 0 || key == "oneOf" && count != 1 {
				return false, nil
			}
		}
	}
	if _, ok := m["not"]; ok {
		ok, e := test(sc.nodes[n.path+"/not"], v)
		if e != nil {
			return false, e
		}
		if ok {
			return false, nil
		}
	}
	bounds := func(size int, lo, hi string) bool {
		x := json.Number(strconv.Itoa(size))
		if b, ok := m[lo].(json.Number); ok && sc.parent.Ensemble().JSON().Compare(x, b) < 0 {
			return false
		}
		if b, ok := m[hi].(json.Number); ok && sc.parent.Ensemble().JSON().Compare(x, b) > 0 {
			return false
		}
		return true
	}
	switch x := v.(type) {
	case map[string]any:
		if !bounds(len(x), "minProperties", "maxProperties") {
			return false, nil
		}
		if a, ok := m["required"].([]any); ok {
			for _, k := range a {
				if _, ok := x[k.(string)]; !ok {
					return false, nil
				}
			}
		}
		props, _ := m["properties"].(map[string]any)
		for k, value := range x {
			var child *schemaNode
			if _, ok := props[k]; ok {
				child = sc.nodes[n.path+"/properties/"+pointerPart(k)]
			} else if _, ok := m["additionalProperties"]; ok {
				child = sc.nodes[n.path+"/additionalProperties"]
			}
			if child != nil {
				ok, e := test(child, value)
				if e != nil || !ok {
					return ok, e
				}
			}
		}
	case []any:
		if !bounds(len(x), "minItems", "maxItems") {
			return false, nil
		}
		if _, ok := m["items"]; ok {
			for _, value := range x {
				ok, e := test(sc.nodes[n.path+"/items"], value)
				if e != nil || !ok {
					return ok, e
				}
			}
		}
	case string:
		if !bounds(utf8.RuneCountInString(x), "minLength", "maxLength") {
			return false, nil
		}
	case json.Number:
		for _, k := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum"} {
			if b, ok := m[k].(json.Number); ok {
				c := sc.parent.Ensemble().JSON().Compare(x, b)
				if k == "minimum" && c < 0 || k == "maximum" && c > 0 || k == "exclusiveMinimum" && c <= 0 || k == "exclusiveMaximum" && c >= 0 {
					return false, nil
				}
			}
		}
	}
	return true, nil
}
func (sc *schema) matches(t string, v any) bool {
	switch t {
	case "null":
		return v == nil
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
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "integer":
		n, ok := v.(json.Number)
		return ok && sc.parent.Ensemble().JSON().Integral(n)
	}
	return false
}
