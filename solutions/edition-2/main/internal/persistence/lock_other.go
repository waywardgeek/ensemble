//go:build !darwin && !linux

package persistence

import (
	"fmt"
	"os"
)

func supported() bool               { return false }
func lockFile(*os.File, bool) error { return fmt.Errorf("unsupported") }
func inUse(error) bool              { return false }
