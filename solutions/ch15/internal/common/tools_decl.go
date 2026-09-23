package common

// EffectiveTools is the tools array for a vendor that cannot carry tool
// declarations in the dialog (Chapter 15 rule 2): the frozen startup set,
// with every KindTools entry's delta applied in dialog order. Such a vendor
// re-declares, and pays the cache miss; a vendor that can carry them renders
// the KindTools entries in place and sends the startup set unchanged.
//
// A change to a tool already declared replaces it in place; a new tool goes
// on the end. Deterministic: the result is a function of the base and the
// context alone.
func EffectiveTools(base []ToolDecl, c *Context) []ToolDecl {
	out := base
	copied := false
	own := func() {
		if !copied {
			out = append([]ToolDecl(nil), out...)
			copied = true
		}
	}
	for _, e := range c.Dialogue {
		if e.Kind != KindTools {
			continue
		}
		for _, p := range e.Parts {
			d, ok := p.(ToolDeclPart)
			if !ok {
				continue
			}
			for _, name := range d.Removed {
				for i := range out {
					if out[i].Name == name {
						own()
						out = append(out[:i], out[i+1:]...)
						break
					}
				}
			}
			for _, t := range d.Added {
				own()
				replaced := false
				for i := range out {
					if out[i].Name == t.Name {
						out[i] = t
						replaced = true
						break
					}
				}
				if !replaced {
					out = append(out, t)
				}
			}
		}
	}
	return out
}
