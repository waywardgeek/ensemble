package main

import (
	gui "ensemble-gui-stub"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func runGUIAuth(args []string, output io.Writer) error {
	if len(args) == 0 || (args[0] != "auth-init" && args[0] != "auth-reset") {
		return errors.New("usage: auth-init|auth-reset [--file path] [--username ensemble] [--config path]")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("file", "cr/gui-auth.json", "private GUI credential file")
	username := flags.String("username", "ensemble", "GUI login username")
	config := flags.String("config", "", "application settings file")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
		return errors.New("invalid GUI authentication arguments; use --file, --username and --config only")
	}
	explicitFile, explicitConfig := false, false
	flags.Visit(func(value *flag.Flag) {
		explicitFile = explicitFile || value.Name == "file"
		explicitConfig = explicitConfig || value.Name == "config"
	})
	if explicitConfig && strings.TrimSpace(*config) == "" {
		return errors.New("--config requires a nonempty path")
	}
	if explicitConfig && !explicitFile {
		application, err := gui.LoadApplicationConfig(*config)
		if err != nil {
			return errors.New("cannot load application settings for GUI authentication")
		}
		*path = filepath.Join(application.Snapshot().DataDir, "gui-auth.json")
	}
	if strings.TrimSpace(*path) == "" {
		return errors.New("--file requires a nonempty path")
	}
	password, err := gui.CreateDigestCredentials(*path, *username, args[0] == "auth-reset")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output, password); err != nil {
		return errors.New("GUI auth credentials saved, but password output failed; use auth-reset")
	}
	return nil
}

func guiAuthPath(host, dataDir, override string, explicit bool) (string, error) {
	path := filepath.Join(dataDir, "gui-auth.json")
	if explicit {
		if strings.TrimSpace(override) == "" {
			return "", errors.New("--gui-auth-file requires a nonempty path")
		}
		path = override
	}
	if _, err := os.Lstat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", errors.New("cannot inspect GUI auth file")
	}
	ip := net.ParseIP(host)
	if !explicit && (host == "localhost" || ip != nil && ip.IsLoopback()) {
		return "", nil
	}
	return "", fmt.Errorf("GUI authentication required; run auth-init --file %q before starting this listener", path)
}

func checkGUIAuthPath(path string, reserved ...string) error {
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return errors.New("cannot inspect GUI auth file")
	}
	for _, other := range reserved {
		if other == "" {
			continue
		}
		// The credential file already exists. Comparing inodes catches relative
		// paths, symlinked directories and hard links before any output is opened.
		otherInfo, statErr := os.Stat(other)
		if statErr == nil && os.SameFile(info, otherInfo) {
			return errors.New("GUI auth file must not use an application log, configuration or persistence path")
		}
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return errors.New("cannot inspect application output path for GUI auth")
		}
	}
	return nil
}
