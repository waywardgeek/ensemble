package gui

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func TestDigestCredentialsPersistence(test *testing.T) {
	directory := test.TempDir()
	path := filepath.Join(directory, "private", "nested", "gui-auth.json")
	password, err := CreateDigestCredentials(path, "ensemble", false)
	if err != nil {
		test.Fatal(err)
	}
	random, err := base64.RawURLEncoding.DecodeString(password)
	if err != nil || len(random) != 16 {
		test.Fatal("generated password must encode 128 random bits as base64url")
	}
	for _, name := range []string{"private", filepath.Join("private", "nested")} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || info.Mode().Perm() != 0700 {
			test.Fatal("new credential directories must be mode 0700")
		}
	}
	before, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	var document map[string]any
	if json.Unmarshal(before, &document) != nil || len(document) != 4 || document["password"] != nil || bytes.Contains(before, []byte(password)) {
		test.Fatal("credential file must contain only the version, username, realm and HA1 verifier")
	}
	want := sha256.Sum256([]byte("ensemble:Ensemble:" + password))
	for attempt := 0; attempt < 2; attempt++ {
		credentials, err := loadDigestCredentials(path)
		if err != nil || credentials.Version != 1 || credentials.Username != "ensemble" || credentials.Realm != "Ensemble" || credentials.HA1 != hex.EncodeToString(want[:]) {
			test.Fatal("reloaded credentials differ from the SHA-256 verifier")
		}
	}
	if _, err := CreateDigestCredentials(path, "other", false); !errors.Is(err, os.ErrExist) {
		test.Fatal("init must refuse an existing credential file")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		test.Fatal("init changed existing credentials")
	}
	oldInfo, err := os.Stat(path)
	if err != nil {
		test.Fatal(err)
	}
	oldFile, err := os.Open(path)
	if err != nil {
		test.Fatal(err)
	}
	defer oldFile.Close()
	replacement, err := CreateDigestCredentials(path, "new-user", true)
	if err != nil || replacement == password || replacement == "" {
		test.Fatal("reset must return a new password")
	}
	newInfo, err := os.Stat(path)
	if err != nil || newInfo.Mode() != 0600 || os.SameFile(oldInfo, newInfo) {
		test.Fatal("reset must atomically replace the private credential file")
	}
	credentials, err := loadDigestCredentials(path)
	want = sha256.Sum256([]byte("new-user:Ensemble:" + replacement))
	if err != nil || credentials.Username != "new-user" || credentials.HA1 != hex.EncodeToString(want[:]) {
		test.Fatal("reset verifier does not match the new login")
	}
	oldData := make([]byte, len(before))
	if _, err := oldFile.Read(oldData); err != nil || !bytes.Equal(oldData, before) {
		test.Fatal("reset modified the old inode in place")
	}
}

func TestDigestCredentialValidation(test *testing.T) {
	directory := test.TempDir()
	validPath := filepath.Join(directory, "valid.json")
	password, err := CreateDigestCredentials(validPath, "ensemble", false)
	if err != nil {
		test.Fatal(err)
	}
	valid, err := os.ReadFile(validPath)
	if err != nil {
		test.Fatal(err)
	}
	var credentials digestCredentials
	if err := json.Unmarshal(valid, &credentials); err != nil {
		test.Fatal("invalid baseline fixture")
	}
	for name, document := range map[string]string{
		"empty":               "",
		"null":                "null",
		"array":               "[]",
		"truncated":           "{",
		"unknown":             strings.Replace(string(valid), "{", `{"password":"`+password+`",`, 1),
		"duplicate":           strings.Replace(string(valid), "{", `{"version":1,`, 1),
		"case alias":          strings.Replace(string(valid), `"version"`, `"Version"`, 1),
		"version":             strings.Replace(string(valid), `"version":1`, `"version":2`, 1),
		"missing version":     strings.Replace(string(valid), `"version":1,`, "", 1),
		"null username":       strings.Replace(string(valid), `"username":"ensemble"`, `"username":null`, 1),
		"username colon":      strings.Replace(string(valid), `"username":"ensemble"`, `"username":"bad:name"`, 1),
		"username control":    strings.Replace(string(valid), `"username":"ensemble"`, `"username":"bad\nname"`, 1),
		"username unicode":    strings.Replace(string(valid), `"username":"ensemble"`, `"username":"caf\u00e9"`, 1),
		"username decomposed": strings.Replace(string(valid), `"username":"ensemble"`, `"username":"cafe\u0301"`, 1),
		"realm":               strings.Replace(string(valid), `"Ensemble"`, `"other"`, 1),
		"uppercase hash":      strings.Replace(string(valid), credentials.HA1, strings.ToUpper(credentials.HA1), 1),
		"short hash":          strings.Replace(string(valid), credentials.HA1, "ab", 1),
		"invalid hash":        strings.Replace(string(valid), credentials.HA1, strings.Repeat("g", 64), 1),
		"extra document":      string(valid) + "{}",
		"trailing garbage":    string(valid) + "garbage",
		"oversized":           string(valid) + strings.Repeat(" ", digestCredentialsLimit),
		"invalid utf8":        strings.Replace(string(valid), "ensemble", string([]byte{0xff}), 1),
	} {
		test.Run(name, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "invalid.json")
			if err := os.WriteFile(path, []byte(document), 0600); err != nil {
				test.Fatal(err)
			}
			if _, err := loadDigestCredentials(path); err == nil {
				test.Fatal("accepted invalid credentials")
			} else if strings.Contains(err.Error(), password) || strings.Contains(err.Error(), credentials.HA1) {
				test.Fatal("credential bytes leaked into error")
			}
			if generated, err := CreateDigestCredentials(path, "ensemble", true); err == nil || generated != "" {
				test.Fatal("reset replaced a noncredential file")
			}
			if after, err := os.ReadFile(path); err != nil || string(after) != document {
				test.Fatal("failed reset changed the destination")
			}
		})
	}
}

func TestDigestCredentialUsernames(test *testing.T) {
	for _, username := range []string{"", "bad:name", "bad name", "bad\nname", "café", "cafe\u0301", "quoted\"name", "back\\slash", strings.Repeat("a", 257)} {
		path := filepath.Join(test.TempDir(), "credentials.json")
		if password, err := CreateDigestCredentials(path, username, false); err == nil || password != "" {
			test.Error("invalid username accepted")
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			test.Error("invalid username created a credential file")
		}
	}
	path := filepath.Join(test.TempDir(), "credentials.json")
	if _, err := CreateDigestCredentials(path, "Operator_2.example-user", false); err != nil {
		test.Fatal("valid ASCII username rejected")
	}
	if _, err := loadDigestCredentials(path); err != nil {
		test.Fatal("valid ASCII username failed to reload")
	}
}

func TestDigestCredentialFileSafety(test *testing.T) {
	for _, kind := range []string{"missing", "directory", "symlink", "dangling symlink", "hardlink", "fifo", "readable", "writable", "readonly", "special mode", "wrong owner"} {
		test.Run(kind, func(test *testing.T) {
			directory := test.TempDir()
			path := filepath.Join(directory, "credentials.json")
			target := filepath.Join(directory, "target.json")
			if _, err := CreateDigestCredentials(target, "ensemble", false); err != nil {
				test.Fatal(err)
			}
			original, err := os.ReadFile(target)
			if err != nil {
				test.Fatal(err)
			}
			switch kind {
			case "missing":
			case "directory":
				err = os.Mkdir(path, 0700)
			case "symlink":
				err = os.Symlink(target, path)
			case "dangling symlink":
				err = os.Symlink(filepath.Join(directory, "absent"), path)
			case "hardlink":
				err = os.Link(target, path)
			case "fifo":
				err = syscall.Mkfifo(path, 0600)
			default:
				err = os.WriteFile(path, original, 0600)
				if err == nil {
					switch kind {
					case "readable":
						err = os.Chmod(path, 0644)
					case "writable":
						err = os.Chmod(path, 0620)
					case "readonly":
						err = os.Chmod(path, 0400)
					case "special mode":
						err = os.Chmod(path, 0600|os.ModeSetuid)
					case "wrong owner":
						if os.Geteuid() != 0 {
							test.Skip("changing file ownership requires root")
						}
						err = os.Chown(path, 1, -1)
					}
				}
			}
			if err != nil {
				test.Fatal(err)
			}
			if _, err := loadDigestCredentials(path); err == nil {
				test.Fatal("unsafe credential file loaded")
			}
			if password, err := CreateDigestCredentials(path, "ensemble", true); err == nil || password != "" {
				test.Fatal("unsafe reset succeeded")
			}
			if kind != "missing" {
				if password, err := CreateDigestCredentials(path, "ensemble", false); err == nil || password != "" {
					test.Fatal("init replaced an existing destination")
				}
			}
			if after, err := os.ReadFile(target); err != nil || !bytes.Equal(after, original) {
				test.Fatal("credential operation modified another file")
			}
		})
	}
}

func TestDigestCredentialsConcurrentInit(test *testing.T) {
	path := filepath.Join(test.TempDir(), "credentials.json")
	results := make(chan string, 8)
	var workers sync.WaitGroup
	for attempt := 0; attempt < cap(results); attempt++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			password, err := CreateDigestCredentials(path, "ensemble", false)
			if err == nil {
				results <- password
			} else if !errors.Is(err, os.ErrExist) {
				test.Error("concurrent init failed for an unexpected reason")
			}
		}()
	}
	workers.Wait()
	close(results)
	if len(results) != 1 {
		test.Fatal("exactly one initializer must publish a password")
	}
	password := <-results
	want := sha256.Sum256([]byte("ensemble:Ensemble:" + password))
	credentials, err := loadDigestCredentials(path)
	if err != nil || credentials.HA1 != hex.EncodeToString(want[:]) {
		test.Fatal("published verifier does not belong to the successful initializer")
	}
}
