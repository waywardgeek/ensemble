package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, data := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
func TestSourceProperties(t *testing.T) {
	for _, test := range []struct {
		name, source string
		want         int
	}{
		{"constant", "package sample\nconst counter = 3\n", 0},
		{"lookup", "package sample\nvar lookup = map[string]int{\"a\":1}\n", 0},
		{"uninitialized-session", "package sample\nvar next uint64\n", 1},
		{"allocated-session", "package sample\nvar inbox = make(chan int)\n", 1},
		{"mutated-session", "package sample\nvar next = 0\nfunc bump(){next++}\n", 1},
		{"different-name", "package sample\nvar citrus = 0\nfunc bump(){citrus++}\n", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := fixture(t, map[string]string{"go.mod": "module fixture.example/lib\n", "owner.go": test.source})
			got, err := inspect(root)
			if err != nil || len(got) != test.want {
				t.Fatalf("got %v, %v; want %d", got, err, test.want)
			}
		})
	}
}
func TestNestedExecutablePrivateImport(t *testing.T) {
	root := fixture(t, map[string]string{"go.mod": "module fixture.example/lib\n", "examples/new/go.mod": "module fixture.example/lib/examples/new\n", "examples/new/main.go": "package main\nimport _ \"fixture.example/lib/internal/newspoke\"\nfunc main(){}\n"})
	got, err := inspect(root)
	if err != nil || len(got) != 1 {
		t.Fatalf("new executable escaped: %v %v", got, err)
	}
}
