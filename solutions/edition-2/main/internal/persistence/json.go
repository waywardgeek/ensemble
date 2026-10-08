// Package persistence owns session encoding and storage, never conversation policy.
package persistence

import (
	"bytes"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strconv"
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
	v, err := c.parseJSON(data, session)
	if err != nil {
		return err
	}
	return c.eventFields(v)
}
func (c *Codec) parseJSON(data []byte, scalar bool) (any, error) {
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
	v, err := c.value(d, 0)
	if err != nil {
		return nil, err
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, c.bad("trailing JSON")
	}
	return v, nil
}

// Check escapes before encoding/json can replace an unpaired surrogate. Skip
// escaped backslashes as a unit: the literal text \\ud800 is not a Unicode escape.
func (c *Codec) scalarEscapes(data []byte) error {
	quoted := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			quoted = !quoted
			continue
		}
		if !quoted || data[i] != '\\' {
			continue
		}
		i++
		if i >= len(data) {
			break
		}
		if data[i] != 'u' {
			continue
		}
		if i+4 >= len(data) {
			return c.bad("invalid Unicode escape")
		}
		n, err := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
		if err != nil {
			return c.bad("invalid Unicode escape")
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return c.bad("unpaired low surrogate")
		}
		if n < 0xd800 || n > 0xdbff {
			continue
		}
		if i+6 >= len(data) || data[i+1] != '\\' || data[i+2] != 'u' {
			return c.bad("unpaired high surrogate")
		}
		low, err := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return c.bad("unpaired high surrogate")
		}
		i += 6
	}
	return nil
}
func (c *Codec) value(d *json.Decoder, depth int) (any, error) {
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
				return nil, c.bad("duplicate JSON member")
			}
			if len(out) >= CollectionLimit {
				return nil, c.bad("object collection limit")
			}
			v, err := c.value(d, depth+1)
			if err != nil {
				return nil, err
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
			v, err := c.value(d, depth+1)
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
func (c *Codec) number(s string) string {
	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	exponent := new(big.Int)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exponent.SetString(strings.TrimPrefix(s[i+1:], "+"), 10)
		s = s[:i]
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		exponent.Sub(exponent, big.NewInt(int64(len(s)-i-1)))
		s = s[:i] + s[i+1:]
	}
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	trimmed := strings.TrimRight(s, "0")
	exponent.Add(exponent, big.NewInt(int64(len(s)-len(trimmed))))
	s = trimmed
	if negative {
		s = "-" + s
	}
	if exponent.Sign() != 0 {
		s += "e" + exponent.String()
	}
	return s
}
func (c *Codec) quoted(b *bytes.Buffer, s string, limit int) error {
	// Check the escaped size before growing a buffer, including a single large string.
	n := 2
	for _, r := range s {
		switch r {
		case '"', '\\', '\b', '\f', '\n', '\r', '\t':
			n += 2
		default:
			if r < 32 {
				n += 6
			} else {
				n += utf8.RuneLen(r)
			}
		}
		if n > limit-b.Len() {
			return c.bad("encoded JSON byte limit")
		}
	}
	if n > limit-b.Len() {
		return c.bad("encoded JSON byte limit")
	}
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 32 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return nil
}
func (c *Codec) canonicalValue(b *bytes.Buffer, v any) error {
	return c.writeJSON(b, v, true, StateLimit)
}
func (c *Codec) writeJSON(b *bytes.Buffer, v any, normalize bool, limit int) error {
	if b.Len() > limit {
		return c.bad("canonical JSON exceeds 256 MiB")
	}
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		return c.quoted(b, x, limit)
	case json.Number:
		number := string(x)
		if normalize {
			number = c.number(number)
		}
		if len(number) > limit-b.Len() {
			return c.bad("encoded JSON byte limit")
		}
		b.WriteString(number)
	case []any:
		b.WriteByte('[')
		for i, v := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := c.writeJSON(b, v, normalize, limit); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := c.quoted(b, k, limit); err != nil {
				return err
			}
			b.WriteByte(':')
			if err := c.writeJSON(b, x[k], normalize, limit); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	default:
		return c.bad("invalid canonical value")
	}
	if b.Len() > limit {
		return c.bad("canonical JSON exceeds 256 MiB")
	}
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
