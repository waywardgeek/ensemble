// Package jsonvalue implements bounded lossless JSON syntax for root-owned callers.
package jsonvalue

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

type Service struct{ parent common.Ensemble }

func New(parent common.Ensemble) *Service    { return &Service{parent: parent} }
func (c *Service) Ensemble() common.Ensemble { return c.parent }
func (c *Service) bad(detail string) error {
	c.parent.Logf("JSON validation: %s", detail)
	return fmt.Errorf("%s", detail)
}

// Check escapes before encoding/json can replace an unpaired surrogate. Skip
// escaped backslashes as a unit: the literal text \\ud800 is not a Unicode escape.
func (c *Service) ScalarEscapes(data []byte) error {
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
func (c *Service) Number(s string) string {
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
func (c *Service) quoted(b *bytes.Buffer, s string, limit int) error {
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
func (c *Service) canonicalValue(b *bytes.Buffer, v any) error {
	return c.writeJSON(b, v, true, 256<<20)
}
func (c *Service) writeJSON(b *bytes.Buffer, v any, normalize bool, limit int) error {
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
			number = c.Number(number)
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

func (c *Service) Encode(v any, normalize bool, limit int) ([]byte, error) {
	if limit < 1 {
		return nil, c.bad("invalid JSON output bound")
	}
	var b bytes.Buffer
	err := c.writeJSON(&b, v, normalize, limit)
	return b.Bytes(), err
}
func (c *Service) Parse(data []byte, b common.JSONBounds) (any, error) {
	if b.Bytes < 1 || b.Depth < 1 || b.Collection < 1 || b.Nodes < 0 || len(data) > b.Bytes {
		return nil, c.bad("JSON byte/profile limit")
	}
	if !utf8.Valid(data) {
		return nil, c.bad("JSON is not UTF-8")
	}
	if b.Scalars {
		if err := c.ScalarEscapes(data); err != nil {
			return nil, err
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	nodes := 0
	v, err := c.value(d, b, 0, &nodes)
	if err != nil {
		return nil, err
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, c.bad("trailing JSON")
	}
	return v, nil
}
func (c *Service) value(d *json.Decoder, b common.JSONBounds, depth int, nodes *int) (any, error) {
	*nodes++
	if b.Nodes > 0 && *nodes > b.Nodes {
		return nil, c.bad("JSON node limit")
	}
	t, err := d.Token()
	if err != nil {
		return nil, c.bad("invalid JSON")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return t, nil
	}
	if depth >= b.Depth {
		return nil, c.bad("JSON depth limit")
	}
	switch delim {
	case '{':
		out := map[string]any{}
		for d.More() {
			if len(out) >= b.Collection {
				return nil, c.bad("JSON collection limit")
			}
			key, err := d.Token()
			if err != nil {
				return nil, c.bad("invalid object key")
			}
			s, ok := key.(string)
			if !ok {
				return nil, c.bad("invalid object key")
			}
			if _, ok := out[s]; ok {
				return nil, c.bad("duplicate JSON member")
			}
			v, err := c.value(d, b, depth+1, nodes)
			if err != nil {
				return nil, err
			}
			out[s] = v
		}
		if t, err = d.Token(); err != nil || t != json.Delim('}') {
			return nil, c.bad("invalid object end")
		}
		return out, nil
	case '[':
		out := []any{}
		for d.More() {
			if len(out) >= b.Collection {
				return nil, c.bad("JSON collection limit")
			}
			v, err := c.value(d, b, depth+1, nodes)
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
	return nil, c.bad("invalid delimiter")
}

// Compare works in coefficient/exponent space; huge exponents never expand.
func (c *Service) decimal(n json.Number) (bool, string, *big.Int) {
	s := c.Number(string(n))
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	e := new(big.Int)
	if i := strings.IndexByte(s, 'e'); i >= 0 {
		e.SetString(s[i+1:], 10)
		s = s[:i]
	}
	return neg, s, e
}
func (c *Service) Integral(n json.Number) bool {
	_, digits, e := c.decimal(n)
	return digits == "0" || e.Sign() >= 0
}
func (c *Service) Compare(a, b json.Number) int {
	an, ad, ae := c.decimal(a)
	bn, bd, be := c.decimal(b)
	if ad == "0" {
		if bd == "0" {
			return 0
		}
		if bn {
			return 1
		}
		return -1
	}
	if bd == "0" {
		if an {
			return -1
		}
		return 1
	}
	if an != bn {
		if an {
			return -1
		}
		return 1
	}
	ae.Add(ae, big.NewInt(int64(len(ad))))
	be.Add(be, big.NewInt(int64(len(bd))))
	cmp := ae.Cmp(be)
	if cmp == 0 {
		n := len(ad)
		if len(bd) > n {
			n = len(bd)
		}
		for i := 0; i < n; i++ {
			x, y := byte('0'), byte('0')
			if i < len(ad) {
				x = ad[i]
			}
			if i < len(bd) {
				y = bd[i]
			}
			if x < y {
				cmp = -1
				break
			}
			if x > y {
				cmp = 1
				break
			}
		}
	}
	if an {
		return -cmp
	}
	return cmp
}
