package main

import (
	"fmt"
	"os"
	"path/filepath"

	"ensemble"
)

// configureSkills keeps launch defaults in this public-API GUI client.
// An explicit environment selection is already present in config.
func configureSkills(config *ensemble.Config, directory string, guiDebug bool) error {
	if config.Skills != nil {
		copy := *config.Skills
		config.Skills = &copy
		if directory != "" {
			config.Skills.Directory = directory
		}
	} else if directory != "" {
		config.Skills = &ensemble.SkillSpec{Directory: directory, Primary: "ensemble"}
	} else {
		path := filepath.Join("skills", "ensemble", "SKILL.md")
		info, err := os.Stat(path)
		switch {
		case err == nil && !info.IsDir():
			config.Skills = &ensemble.SkillSpec{Directory: "skills", Primary: "ensemble"}
		case err != nil && !os.IsNotExist(err):
			return fmt.Errorf("default primary skill: %w", err)
		case guiDebug:
			return fmt.Errorf("--gui-debug requires --skills-dir or %s", path)
		}
	}
	if guiDebug {
		for _, name := range config.StartupSkills {
			if name == "gui-debug" {
				return nil
			}
		}
		config.StartupSkills = append(config.StartupSkills, "gui-debug")
	}
	return nil
}
