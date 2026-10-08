package persistence

import (
	"crypto/sha256"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func (c *Codec) fieldName(f reflect.StructField) string {
	name := strings.Split(f.Tag.Get("json"), ",")[0]
	if name == "" {
		name = f.Name
	}
	return name
}
func (c *Codec) wire(v reflect.Value, semantic bool) (any, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		return c.wire(v.Elem(), semantic)
	}
	if v.Type() == reflect.TypeOf(json.RawMessage{}) {
		if v.IsNil() {
			return nil, nil
		}
		raw := v.Bytes()
		value, err := c.parse(raw)
		if err != nil {
			return nil, err
		}
		if semantic {
			return value, nil
		}
		return string(raw), nil
	}
	switch v.Kind() {
	case reflect.Struct:
		out := map[string]any{}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name := c.fieldName(f)
			if name == "-" {
				continue
			}
			if t == reflect.TypeOf(common.LimitValues{}) && v.Field(i).IsNil() {
				continue
			}
			value, err := c.wire(v.Field(i), t == reflect.TypeOf(common.HandlerIdentity{}) && name == "schema")
			if err != nil {
				return nil, err
			}
			out[name] = value
		}
		return out, nil
	case reflect.Slice:
		out := make([]any, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			x, err := c.wire(v.Index(i), false)
			if err != nil {
				return nil, err
			}
			out = append(out, x)
		}
		return out, nil
	case reflect.Map:
		out := map[string]any{}
		it := v.MapRange()
		for it.Next() {
			key := fmt.Sprint(it.Key().Interface())
			x, err := c.wire(it.Value(), false)
			if err != nil {
				return nil, err
			}
			out[key] = x
		}
		return out, nil
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		return json.Number(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Int, reflect.Int64, reflect.Int32:
		return json.Number(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Float64:
		if math.IsNaN(v.Float()) || math.IsInf(v.Float(), 0) {
			return nil, c.bad("invalid number")
		}
		return json.Number(strconv.FormatFloat(v.Float(), 'g', -1, 64)), nil
	}
	return nil, c.bad("unsupported semantic field")
}
func (c *Codec) unwire(x any, v reflect.Value, semantic bool) error {
	if v.Kind() == reflect.Pointer {
		if x == nil {
			v.SetZero()
			return nil
		}
		v.Set(reflect.New(v.Type().Elem()))
		return c.unwire(x, v.Elem(), semantic)
	}
	if v.Type() == reflect.TypeOf(json.RawMessage{}) {
		if x == nil {
			v.SetZero()
			return nil
		}
		if semantic {
			raw, err := json.Marshal(x)
			if err != nil {
				return c.bad("invalid schema")
			}
			v.SetBytes(raw)
			return nil
		}
		s, ok := x.(string)
		if !ok {
			return c.bad("raw JSON must be a string")
		}
		if _, err := c.parse([]byte(s)); err != nil {
			return err
		}
		v.SetBytes([]byte(s))
		return nil
	}
	if x == nil {
		return c.bad("null required semantic field")
	}
	switch v.Kind() {
	case reflect.Struct:
		obj, ok := x.(map[string]any)
		if !ok {
			return c.bad("expected structural object")
		}
		t := v.Type()
		names := map[string]bool{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name := c.fieldName(f)
			if name == "-" {
				continue
			}
			names[name] = true
			value, found := obj[name]
			if !found {
				if t == reflect.TypeOf(common.LimitValues{}) {
					continue
				}
				return c.bad("missing structural field " + name)
			}
			if t == reflect.TypeOf(common.LimitValues{}) && value == nil {
				return c.bad("null limit override")
			}
			if err := c.unwire(value, v.Field(i), t == reflect.TypeOf(common.HandlerIdentity{}) && name == "schema"); err != nil {
				return err
			}
		}
		for name := range obj {
			if !names[name] {
				return c.bad("unknown structural field")
			}
		}
		return nil
	case reflect.Slice:
		list, ok := x.([]any)
		if !ok {
			return c.bad("expected array")
		}
		if len(list) > CollectionLimit {
			return c.bad("collection limit")
		}
		v.Set(reflect.MakeSlice(v.Type(), len(list), len(list)))
		for i, x := range list {
			if err := c.unwire(x, v.Index(i), false); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		obj, ok := x.(map[string]any)
		if !ok {
			return c.bad("expected map")
		}
		v.Set(reflect.MakeMap(v.Type()))
		for key, x := range obj {
			k := reflect.New(v.Type().Key()).Elem()
			if k.Kind() == reflect.String {
				k.SetString(key)
			} else {
				n, err := strconv.ParseUint(key, 10, 64)
				if err != nil || strconv.FormatUint(n, 10) != key || n == 0 {
					return c.bad("invalid identity map key")
				}
				k.SetUint(n)
			}
			value := reflect.New(v.Type().Elem()).Elem()
			if err := c.unwire(x, value, false); err != nil {
				return err
			}
			v.SetMapIndex(k, value)
		}
		return nil
	case reflect.String:
		s, ok := x.(string)
		if !ok {
			return c.bad("expected string")
		}
		v.SetString(s)
		return nil
	case reflect.Bool:
		b, ok := x.(bool)
		if !ok {
			return c.bad("expected Boolean")
		}
		v.SetBool(b)
		return nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		n, ok := x.(json.Number)
		if !ok {
			return c.bad("expected uint64")
		}
		u, err := strconv.ParseUint(string(n), 10, v.Type().Bits())
		if err != nil {
			return c.bad("invalid uint64")
		}
		v.SetUint(u)
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		n, ok := x.(json.Number)
		if !ok {
			return c.bad("expected integer")
		}
		i, err := strconv.ParseInt(string(n), 10, v.Type().Bits())
		if err != nil {
			return c.bad("invalid integer")
		}
		v.SetInt(i)
		return nil
	case reflect.Float64:
		n, ok := x.(json.Number)
		if !ok {
			return c.bad("expected number")
		}
		f, err := strconv.ParseFloat(string(n), 64)
		if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
			return c.bad("invalid number")
		}
		v.SetFloat(f)
		return nil
	}
	return c.bad("unsupported semantic field")
}
func (c *Codec) Encode(cp common.Checkpoint) ([]byte, error) {
	// Generic struct serialization is used only after conversion to the strict
	// wire grammar; raw fields have become strings and own their original bytes.
	root, err := c.wire(reflect.ValueOf(cp), false)
	if err != nil {
		return nil, err
	}
	obj := root.(map[string]any)
	obj["state"] = nil
	obj["state_sha256"] = nil
	if cp.State != nil {
		state, err := c.wire(reflect.ValueOf(*cp.State), false)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(state)
		if err != nil {
			return nil, c.bad("cannot encode state")
		}
		canonical, err := c.Canonical(raw)
		if err != nil {
			return nil, &common.SessionError{Code: "session_limit", Detail: "canonical state limit or invalid state"}
		}
		obj["state"] = state
		obj["state_sha256"] = fmt.Sprintf("%x", sha256.Sum256(canonical))
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, c.bad("cannot encode checkpoint")
	}
	if len(raw)+1 > FileLimit {
		return nil, &common.SessionError{Code: "session_limit", Detail: "checkpoint file exceeds 512 MiB"}
	}
	return append(raw, '\n'), nil
}
func (c *Codec) Decode(raw []byte) (common.Checkpoint, error) {
	var cp common.Checkpoint
	if len(raw) > FileLimit {
		return cp, c.bad("checkpoint exceeds 512 MiB")
	}
	value, err := c.parse(raw)
	if err != nil {
		return cp, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return cp, c.bad("checkpoint must be an object")
	}
	state, found := obj["state"]
	if !found {
		return cp, c.bad("missing state")
	}
	delete(obj, "state")
	if err = c.unwire(obj, reflect.ValueOf(&cp).Elem(), false); err != nil {
		return cp, err
	}
	if cp.Version != 1 || cp.StateVersion != 1 {
		return cp, c.bad("unsupported checkpoint version")
	}
	if cp.AsOf == 0 || cp.HighWatermarks.Event != cp.AsOf || !hexadecimal(cp.SessionID, 32) {
		return cp, c.bad("invalid checkpoint boundary")
	}
	if err = c.Identity(cp.Identity); err != nil {
		return cp, err
	}
	if state == nil {
		if cp.StateSHA256 != nil {
			return cp, c.bad("null state hash mismatch")
		}
		return cp, nil
	}
	if cp.StateSHA256 == nil || !hexadecimal(*cp.StateSHA256, 64) {
		return cp, c.bad("invalid state hash")
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return cp, c.bad("invalid state")
	}
	canonical, err := c.Canonical(encoded)
	if err != nil {
		return cp, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(canonical)) != *cp.StateSHA256 {
		return cp, c.bad("state hash mismatch")
	}
	cp.State = &common.SemanticState{}
	if err = c.unwire(state, reflect.ValueOf(cp.State).Elem(), false); err != nil {
		return cp, err
	}
	s := cp.State.Session
	x, _ := json.Marshal(s.Identity)
	y, _ := json.Marshal(cp.Identity)
	if s.ID != cp.SessionID || s.AsOf != cp.AsOf || s.HighWatermarks != cp.HighWatermarks || !c.EqualJSON(x, y) {
		return cp, c.bad("state metadata mismatch")
	}
	return cp, nil
}
func hexadecimal(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func (c *Codec) Identity(id common.SessionIdentity) error {
	if id.Mode == "plain" {
		if id.System == nil || *id.System == "" || id.Skills != nil {
			return c.bad("invalid plain identity")
		}
	} else if id.Mode == "skills" {
		if id.System != nil || id.Skills == nil || id.Skills.Primary == "" || !hexadecimal(id.Skills.CatalogSHA256, 64) || !hexadecimal(id.Skills.BindingsSHA256, 64) {
			return c.bad("invalid skills identity")
		}
	} else {
		return c.bad("invalid identity mode")
	}
	if id.Handlers == nil || len(id.Handlers) > 1024 {
		return c.bad("handler definition count")
	}
	for i, h := range id.Handlers {
		if h.Name == "" || h.Description == "" || i > 0 && id.Handlers[i-1].Name >= h.Name {
			return c.bad("invalid handler identity")
		}
		v, err := c.parse(h.Schema)
		if err != nil {
			return err
		}
		if _, ok := v.(map[string]any); !ok {
			return c.bad("schema must be an object")
		}
	}
	raw, _ := json.Marshal(id.Handlers)
	canonical, err := c.Canonical(raw)
	if err != nil {
		return err
	}
	if len(canonical) > 16<<20 {
		return c.bad("handler definition byte limit")
	}
	return nil
}
