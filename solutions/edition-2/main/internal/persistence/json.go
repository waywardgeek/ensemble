// Package persistence owns session encoding and storage, never conversation policy.
package persistence

import (
	"bytes"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"io"
	"strings"
	"unicode/utf8"
)

const FileLimit = 512 << 20
const StateLimit = 256 << 20
const CollectionLimit = 1000000

type Codec struct{ parent common.Agent }

func NewCodec(parent common.Agent) *Codec { return &Codec{parent: parent} }
func (c *Codec) Agent() common.Agent      { return c.parent }
func (c *Codec) bad(detail string) error {
	c.parent.Ensemble().Logf("session validation: %s", detail)
	return &common.SessionError{Code: "session_corrupt", Detail: detail}
}

// Parse checks duplicate members before maps can discard them. UseNumber keeps
// arbitrary payload integers and decimal exponents out of binary floating point.
func (c *Codec) parse(data []byte) (any, error) {
	return c.parseJSON(data, true)
}
func (c *Codec) ValidateLogJSON(data []byte, session bool) error {
	v, err := c.parseMode(data, session, true, false, nil)
	if err != nil {
		return err
	}
	return c.eventFields(v)
}
func (c *Codec) parseJSON(data []byte, scalar bool) (any, error) {
	v, err := c.parent.Ensemble().JSON().Parse(data, common.JSONBounds{Bytes: FileLimit, Depth: 128, Collection: CollectionLimit, Scalars: scalar})
	if err != nil {
		return nil, c.bad(err.Error())
	}
	return v, nil
}
func (c *Codec) parseMode(data []byte, scalar, event, arguments bool, duplicate *bool) (any, error) {
	if !utf8.Valid(data) {
		return nil, c.bad("JSON is not UTF-8")
	}
	if scalar {
		if err := c.scalarEscapes(data); err != nil {
			return nil, err
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	v, err := c.value(d, 0, "", event, arguments, duplicate)
	if err != nil {
		return nil, err
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, c.bad("trailing JSON")
	}
	return v, nil
}

func (c *Codec) scalarEscapes(data []byte) error {
	if err := c.parent.Ensemble().JSON().ScalarEscapes(data); err != nil {
		return c.bad(err.Error())
	}
	return nil
}
func (c *Codec) value(d *json.Decoder, depth int, path string, event, arguments bool, duplicate *bool) (any, error) {
	t, err := d.Token()
	if err != nil {
		return nil, c.bad("invalid JSON")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return t, nil
	}
	if depth >= 128 {
		return nil, c.bad("JSON nesting exceeds 128")
	}
	switch delim {
	case '{':
		out := map[string]any{}
		count := 0
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return nil, c.bad("invalid object key")
			}
			key, ok := t.(string)
			if !ok {
				return nil, c.bad("invalid object key")
			}
			if _, ok = out[key]; ok {
				if !arguments {
					return nil, c.bad("duplicate JSON member")
				}
				if duplicate != nil {
					*duplicate = true
				}
			}
			count++
			if count > CollectionLimit {
				return nil, c.bad("object collection limit")
			}
			child := path + "/" + key
			argumentField := event && c.argumentPath(child)
			v, err := c.value(d, depth+1, child, event, arguments || argumentField, duplicate)
			if err != nil {
				return nil, err
			}
			if argumentField {
				if _, ok := v.(map[string]any); !ok {
					return nil, c.bad("arguments must be an object")
				}
			}
			out[key] = v
		}
		if t, err = d.Token(); err != nil || t != json.Delim('}') {
			return nil, c.bad("invalid object end")
		}
		return out, nil
	case '[':
		out := []any{}
		for d.More() {
			if len(out) >= CollectionLimit {
				return nil, c.bad("array collection limit")
			}
			v, err := c.value(d, depth+1, path+"/*", event, arguments, duplicate)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		if t, err = d.Token(); err != nil || t != json.Delim(']') {
			return nil, c.bad("invalid array end")
		}
		return out, nil
	}
	return nil, c.bad("invalid JSON delimiter")
}
func (c *Codec) canonicalValue(b *bytes.Buffer, v any) error {
	return c.writeJSON(b, v, true, StateLimit)
}
func (c *Codec) writeJSON(b *bytes.Buffer, v any, normalize bool, limit int) error {
	data, err := c.parent.Ensemble().JSON().Encode(v, normalize, limit-b.Len())
	if err != nil {
		return c.bad(err.Error())
	}
	b.Write(data)
	return nil
}
func (c *Codec) Canonical(data []byte) ([]byte, error) {
	if len(data) > FileLimit {
		return nil, c.bad("JSON file exceeds 512 MiB")
	}
	v, err := c.parse(data)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = c.canonicalValue(&b, v)
	return b.Bytes(), err
}
func (c *Codec) EqualJSON(a, b []byte) bool {
	x, e := c.Canonical(a)
	if e != nil {
		return false
	}
	y, e := c.Canonical(b)
	return e == nil && bytes.Equal(x, y)
}

// Only typed event argument positions may contain ambiguous JSON members.
// All surrounding structures (including the args member itself) remain strict.
func (c *Codec) argumentPath(path string) bool {
	if path == "/tool/args" {
		return true
	}
	for _, prefix := range []string{"/response/parts/*", "/message/parts/*", "/tool/parts/*"} {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		rest := strings.TrimPrefix(path, prefix)
		for strings.HasPrefix(rest, "/parts/*") {
			rest = strings.TrimPrefix(rest, "/parts/*")
		}
		if rest == "/args" {
			return true
		}
	}
	return false
}
func (c *Codec) arguments(raw []byte, scalar bool) (any, bool, error) {
	if len(raw) > common.SkillRecordLimit {
		return nil, false, c.bad("argument byte limit")
	}
	duplicate := false
	value, err := c.parseMode(raw, scalar, false, true, &duplicate)
	if err != nil {
		return nil, false, err
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, false, c.bad("arguments must be an object")
	}
	return value, duplicate, nil
}

// Ambiguous objects have no canonical map meaning. Never discard their members
// to compare them; exact accepted text is the only correspondence in that case.
func (c *Codec) EqualArguments(a, b []byte, scalar bool) bool {
	x, xd, err := c.arguments(a, scalar)
	if err != nil {
		return false
	}
	y, yd, err := c.arguments(b, scalar)
	if err != nil {
		return false
	}
	if xd || yd {
		return bytes.Equal(a, b)
	}
	var left, right bytes.Buffer
	return c.canonicalValue(&left, x) == nil && c.canonicalValue(&right, y) == nil && bytes.Equal(left.Bytes(), right.Bytes())
}
