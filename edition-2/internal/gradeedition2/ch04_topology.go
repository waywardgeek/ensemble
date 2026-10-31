package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// go list supplies resolved paths, not a guessed module name from imports.
// In particular, a consumer must use THIS submission, not an installed copy.
type listedPackage struct {
	// Standard distinguishes Go's standard library from dot-free module names.
	Standard bool
	// ImportPath identifies this package in the resolved dependency graph.
	ImportPath string
	// Imports are direct edges, needed to distinguish public from internal use.
	Imports []string
	// Module records the actual filesystem origin selected by Go's resolver.
	Module *struct {
		// Path is the module's declared import prefix.
		Path string
		// Dir is the resolved module root, including local replacements.
		Dir string
	}
}

func listPackages(dir string, args ...string) ([]listedPackage, error) {
	cmd := exec.Command("go", append([]string{"list", "-json"}, args...)...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list in %s: %w: %s", dir, err, &stderr)
	}
	var packages []listedPackage
	decoder := json.NewDecoder(bytes.NewReader(data))
	// go list writes a stream of package objects, not one JSON array. A failed
	// command is not an empty graph: it must never earn an absence-based pass.
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			return packages, nil
		} else if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
}

func consumerTopology(core, consumer string) grade.Check {
	c := grade.Check{ID: "star-topology", Title: "Star imports and an external consumer of this submission's public package", Points: 25, Earned: 25, Passed: true}
	coreDir, err := filepath.EvalSymlinks(core)
	if err == nil {
		coreDir, err = filepath.Abs(coreDir)
	}
	if err != nil {
		fail(&c, "core directory: %v", err)
		return c
	}
	base := grade.DiscoverBase(coreDir)
	consumerBase := grade.DiscoverBase(consumer)
	// A nested directory alone does not establish an external application. Its
	// own module also prevents this check from grading an internal test helper.
	if base == "" || consumerBase == "" || base == consumerBase {
		fail(&c, "core and consumer need distinct module declarations")
		return c
	}
	packages, err := listPackages(coreDir, "./...")
	if err != nil {
		fail(&c, "%v", err)
		return c
	}
	common, internal := base+"/internal/common", base+"/internal/"
	foundCommon := false
	// Bill's binding common-package rule identifies the hub. Inferring the hub
	// from the most popular import can bless the wrong node after a bad edge.
	// Inspect common's edges too; the old check skipped the discovered hub.
	for _, pkg := range packages {
		if pkg.ImportPath == common {
			foundCommon = true
		}
		if !strings.HasPrefix(pkg.ImportPath, internal) {
			continue // The public composition root intentionally imports spokes.
		}
		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, internal) && imp != common {
				fail(&c, "%s imports implementation package %s", pkg.ImportPath, imp)
			}
		}
	}
	if !foundCommon {
		fail(&c, "required shared vocabulary package internal/common is absent")
	}
	if err := publicDependency(coreDir, consumer); err != nil {
		fail(&c, "%v", err)
	}
	c.Details = append(c.Details, "source review must verify real composition-root use, parent interfaces, common-package behavior and absence of mutable application globals")
	return c
}

// Both external exercises must execute the submitted framework. Share this
// boundary check rather than let each application validate a different checkout.
func publicDependency(core, consumer string) error {
	return publicDependencyTarget(core, consumer, ".")
}

// Check the package actually built: a clean library does not excuse private
// imports or another core checkout in its executable composition root.
func publicDependencyTarget(core, consumer, target string) error {
	coreDir, err := filepath.EvalSymlinks(core)
	if err == nil {
		coreDir, err = filepath.Abs(coreDir)
	}
	if err != nil {
		return err
	}
	base, consumerBase := grade.DiscoverBase(coreDir), grade.DiscoverBase(consumer)
	if base == "" || consumerBase == "" || base == consumerBase {
		return fmt.Errorf("core and consumer need distinct module declarations")
	}
	dependencies, err := listPackages(consumer, "-deps", target)
	if err != nil {
		return err
	}
	usesPublic := false
	for _, pkg := range dependencies {
		if pkg.Module == nil {
			continue
		}
		if pkg.Module.Path == base {
			// A replace directive may resolve the right import name to an older
			// checkout. Compare canonical directories, not just module spelling.
			actual, err := filepath.EvalSymlinks(pkg.Module.Dir)
			if err != nil || actual != coreDir {
				return fmt.Errorf("consumer resolves %s to %s, not submission %s", base, pkg.Module.Dir, coreDir)
			}
		}
		if pkg.Module.Path != consumerBase {
			continue
		}
		// Internal packages rightly appear as transitive dependencies of the
		// public root. Reject only the consumer's DIRECT private imports here.
		// Go's own internal rule is weaker when module prefixes are related.
		for _, imp := range pkg.Imports {
			if imp != base && !strings.HasPrefix(imp, base+"/") {
				continue
			}
			if strings.Contains(imp, "/internal/") {
				return fmt.Errorf("consumer imports private package %s", imp)
			} else {
				usesPublic = true
			}
		}
	}
	if !usesPublic {
		// Pair the absence of forbidden edges with a positive public edge. An
		// unrelated executable must not pass merely because it imports no core.
		return fmt.Errorf("consumer does not import a public package from the submitted core")
	}
	return nil
}
