package grade

import (
	"context"
	"encoding/json"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// The driver runs under node, loads the student's own browser modules against a
// stubbed speechSynthesis, and reports what reached the channel. Keeping it in a
// real .js file rather than a Go string literal means it can be run by hand
// during development, which is how it was built.
//
//go:embed ch14_driver.js
var ch14DriverJS string

// Ch14Result is the raw observation set. One OK/Err pair per behavioural check,
// matching the house pattern from chapters 12 and 13.
type Ch14Result struct {
	BuffersOK  bool
	BuffersErr string

	FiltersOK  bool
	FiltersErr string

	BoundariesOK  bool
	BoundariesErr string

	UnstreamedOK  bool
	UnstreamedErr string

	IdentifiersOK  bool
	IdentifiersErr string

	GateOK  bool
	GateErr string

	Ch13Parity    bool
	Ch13ParityErr string
}

// ch14Spec is the argument handed to the driver: which of the student's files
// hold which role, and what each one calls itself.
type ch14Spec struct {
	TTSFile        string `json:"ttsFile"`
	TTSGlobal      string `json:"ttsGlobal"`
	ArtifactFile   string `json:"artifactFile"`
	ArtifactGlobal string `json:"artifactGlobal"`
	GUIFile        string `json:"guiFile"`
}

type ch14DriverOut struct {
	Checks map[string]struct {
		OK  bool   `json:"ok"`
		Err string `json:"err"`
	} `json:"checks"`
}

// Ch14Run loads the student's speech pipeline under node and exercises it.
func Ch14Run(dir string) Ch14Result {
	var r Ch14Result

	node, err := exec.LookPath("node")
	if err != nil {
		return ch14Fail(fmt.Sprintf("prerequisite missing: node was not found on PATH. "+
			"Chapter 14 grades a browser module, so the grader runs it under node. Install node and re-run. (%v)", err))
	}

	spec, err := ch14Discover(dir)
	if err != nil {
		return ch14Fail(err.Error())
	}
	work, err := os.MkdirTemp("", "ch14grade")
	if err != nil {
		return ch14Fail(fmt.Sprintf("could not create a temp dir: %v", err))
	}
	defer os.RemoveAll(work)

	driver := filepath.Join(work, "ch14_driver.js")
	if err := os.WriteFile(driver, []byte(ch14DriverJS), 0o644); err != nil {
		return ch14Fail(fmt.Sprintf("could not write the driver: %v", err))
	}

	blob, err := json.Marshal(spec)
	if err != nil {
		return ch14Fail(fmt.Sprintf("could not encode the driver spec: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, node, driver, string(blob))
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return ch14Fail(fmt.Sprintf("the driver did not complete (%v). stderr: %s",
			err, ch14Tail(stderr.String())))
	}

	var out ch14DriverOut
	if err := json.Unmarshal([]byte(stdout.String()), &out); err != nil {
		return ch14Fail(fmt.Sprintf("could not parse driver output (%v). stdout: %s stderr: %s",
			err, ch14Tail(stdout.String()), ch14Tail(stderr.String())))
	}

	get := func(id string) (bool, string) {
		c, ok := out.Checks[id]
		if !ok {
			return false, "the driver reported no result for " + id
		}
		return c.OK, c.Err
	}

	r.BuffersOK, r.BuffersErr = get("tts-buffers-fragments")
	r.FiltersOK, r.FiltersErr = get("tts-filters-markup")
	r.BoundariesOK, r.BoundariesErr = get("tts-boundaries")
	r.UnstreamedOK, r.UnstreamedErr = get("tts-speaks-unstreamed")
	r.IdentifiersOK, r.IdentifiersErr = get("tts-expands-identifiers")
	r.GateOK, r.GateErr = get("tts-gate-both-causes")
	return r
}

// ch14Fail marks every behavioural check with the same setup failure, so the
// report says what went wrong rather than reporting six mysterious zeroes.
func ch14Fail(msg string) Ch14Result {
	return Ch14Result{
		BuffersErr:     msg,
		FiltersErr:     msg,
		BoundariesErr:  msg,
		UnstreamedErr:  msg,
		IdentifiersErr: msg,
		GateErr:        msg,
	}
}

func ch14Tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 600 {
		return "..." + s[len(s)-600:]
	}
	return s
}

var (
	ch14ObjDecl   = regexp.MustCompile(`(?m)^(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*\{`)
	ch14ClassDecl = regexp.MustCompile(`(?m)^(?:class\s+([A-Za-z_$][\w$]*)|(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*class)`)
)

// ch14Discover finds the student's browser modules by what they contain rather
// than by what they are called, and reads each module's chosen name out of its
// own source. A student who renames the files or the objects still grades.
func ch14Discover(root string) (ch14Spec, error) {
	var spec ch14Spec
	var files []string

	// The driver runs with its working directory set to the student tree, so every
	// path handed to it must be absolute.
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "testdata", "solutions":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".js") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return spec, fmt.Errorf("could not scan %s for JavaScript files: %v", root, err)
	}
	if len(files) == 0 {
		return spec, fmt.Errorf("found no .js files under %s; chapter 14 grades the browser-side speech pipeline", root)
	}

	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		src := string(b)

		// The speech pipeline is the module with a chunk entry point and a flush.
		if spec.TTSFile == "" && strings.Contains(src, "queueChunk") && strings.Contains(src, "flush") &&
			strings.Contains(src, "SpeechSynthesisUtterance") {
			if m := ch14ObjDecl.FindStringSubmatch(src); m != nil {
				spec.TTSFile, spec.TTSGlobal = p, m[1]
			}
		}

		// The artifact stream is the module that dispatches part_final and feeds speech.
		if spec.ArtifactFile == "" && strings.Contains(src, "part_final") && strings.Contains(src, "handleMessage") {
			if m := ch14ClassDecl.FindStringSubmatch(src); m != nil {
				name := m[1]
				if name == "" {
					name = m[2]
				}
				spec.ArtifactFile, spec.ArtifactGlobal = p, name
			}
		}

		// The pause gate is wherever unpause is sent.
		if spec.GUIFile == "" && strings.Contains(src, "unpause") && !strings.Contains(src, "queueChunk(") {
			spec.GUIFile = p
		}
	}

	if spec.TTSFile == "" {
		return spec, fmt.Errorf("could not find the speech pipeline under %s: looked for a .js module declaring an object with a chunk entry point (queueChunk), a flush(), and a SpeechSynthesisUtterance", root)
	}
	return spec, nil
}
