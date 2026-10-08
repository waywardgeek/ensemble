package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImmutableDeclarations(t *testing.T) {
	cases := []struct{ name, declaration, extra, other, want string }{
		{"errors literal", `import "errors"; var Sentinel = errors.New("stop")`, "", "", ""},
		{"renamed and aliased", `import e "errors"; var CompletelyDifferent = e.New("stop")`, "", "", ""},
		{"formatted literal", `import f "fmt"; var Sentinel = f.Errorf("stop")`, "", "", ""},
		{"embedded renamed", "import e \"embed\"\n//go:embed page.txt\nvar Resources e.FS", "", "", ""},
		{"local shadow", `import "errors"; var Sentinel = errors.New("stop")`, `func f(){ Sentinel := 1; Sentinel++; _ = Sentinel }`, "", ""},
		{"same file assignment", `import "errors"; var Sentinel = errors.New("stop")`, `func f(){ Sentinel = nil }`, "", "runtime mutable value: Sentinel"},
		{"cross file assignment", `import "errors"; var Sentinel = errors.New("stop")`, "", "package fixture\nfunc f(){ Sentinel = nil }", "runtime mutable value: Sentinel"},
		{"address escape", `import "errors"; var Sentinel = errors.New("stop")`, `func f() any { return &Sentinel }`, "", "runtime mutable value: Sentinel"},
		{"embed assignment", "import \"embed\"\n//go:embed page.txt\nvar Resources embed.FS", `func f(){ Resources = embed.FS{} }`, "", "mutable zero value: Resources"},
		{"embed needs directive", `import "embed"; var Resources embed.FS`, "", "", "mutable zero value: Resources"},
		{"dynamic formatting", `import "fmt"; var Sentinel = fmt.Errorf("%s", "stop")`, "", "", "runtime mutable value: Sentinel"},
		{"retained channel rejection", `var Work = make(chan int)`, "", "", "runtime mutable value: Work"},
		{"retained mutable counter", `var Count = 0`, `func f(){ Count++ }`, "", "runtime mutable value: Count"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			write("go.mod", "module fixture\ngo 1.25\n")
			write("main.go", "package fixture\n"+c.declaration+"\n"+c.extra)
			write("page.txt", "hello")
			if c.other != "" {
				write("other.go", c.other)
			}
			got, err := inspect(root)
			if err != nil {
				t.Fatal(err)
			}
			if c.want == "" {
				if len(got) != 0 {
					t.Fatalf("valid immutable declaration rejected: %+v", got)
				}
				return
			}
			if len(got) != 1 || !strings.Contains(got[0].Why, c.want) {
				t.Fatalf("wanted specific %q, got %+v", c.want, got)
			}
		})
	}
}
func TestImportedAssignment(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"go.mod": "module fixture\ngo 1.25\n", "owner/owner.go": "package owner\nimport \"errors\"\nvar Renamed = errors.New(\"stop\")", "client/client.go": "package client\nimport o \"fixture/owner\"\nfunc f(){o.Renamed=nil}"}
	for path, body := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.Contains(got[0].Why, "runtime mutable value: Renamed") {
		t.Fatalf("imported write not rejected: %+v", got)
	}
}
