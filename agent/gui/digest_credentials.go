package gui

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
)

const digestCredentialsLimit = 4096

type digestCredentials struct {
	// Version identifies the on-disk verifier format, currently 1.
	Version int `json:"version"`
	// Username is the login identity included in the SHA-256 input.
	Username string `json:"username"`
	// Realm binds the verifier to the Ensemble authentication challenge.
	Realm string `json:"realm"`
	// HA1 is lowercase SHA-256 hex of username:realm:password, never a password.
	HA1 string `json:"ha1"`
}

// CreateDigestCredentials explicitly creates or resets the GUI's SHA-256
// verifier. The returned random password is never stored; callers must deliver
// it to the user once. Reset requires an existing, valid private credential file.
func CreateDigestCredentials(path, username string, replace bool) (string, error) {
	if !validDigestUsername(username) {
		return "", errors.New("GUI auth username must be 1-256 ASCII letters, digits, dots, underscores or hyphens")
	}
	var original os.FileInfo
	if replace {
		var err error
		original, err = digestCredentialInfo(path)
		if err != nil {
			return "", err
		}
		if _, err := loadDigestCredentials(path); err != nil {
			return "", err
		}
	} else if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("GUI auth file already exists: %w", os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", errors.New("cannot inspect GUI auth file")
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		return "", errors.New("cannot generate GUI auth password")
	}
	password := base64.RawURLEncoding.EncodeToString(secret)
	digest := sha256.Sum256([]byte(username + ":Ensemble:" + password))
	credentials := digestCredentials{Version: 1, Username: username, Realm: "Ensemble", HA1: hex.EncodeToString(digest[:])}
	data, err := json.Marshal(credentials)
	if err != nil {
		return "", errors.New("cannot encode GUI auth credentials")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", errors.New("cannot create GUI auth directory")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".gui-auth-*")
	if err != nil {
		return "", errors.New("cannot create temporary GUI auth file")
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return "", errors.New("cannot secure GUI auth file")
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return "", errors.New("cannot write GUI auth file")
	}
	if err := file.Sync(); err != nil {
		return "", errors.New("cannot sync GUI auth file")
	}
	if err := file.Close(); err != nil {
		return "", errors.New("cannot close GUI auth file")
	}
	if replace {
		current, err := digestCredentialInfo(path)
		if err != nil || !os.SameFile(original, current) {
			return "", errors.New("GUI auth file changed during reset")
		}
		if err := os.Rename(file.Name(), path); err != nil {
			return "", errors.New("cannot replace GUI auth file")
		}
	} else {
		// Publishing a complete inode with Link is atomic and cannot overwrite a
		// file created by another initializer between our existence check and now.
		if err := os.Link(file.Name(), path); err != nil {
			if errors.Is(err, os.ErrExist) {
				return "", fmt.Errorf("GUI auth file already exists: %w", os.ErrExist)
			}
			return "", errors.New("cannot publish GUI auth file")
		}
	}
	return password, nil
}

func digestCredentialInfo(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("GUI auth file is missing: %w", os.ErrNotExist)
		}
		return nil, errors.New("cannot inspect GUI auth file")
	}
	if err := validateDigestCredentialInfo(info); err != nil {
		return nil, err
	}
	return info, nil
}

func validateDigestCredentialInfo(info os.FileInfo) error {
	if !info.Mode().IsRegular() || info.Mode() != 0600 {
		return errors.New("GUI auth file must be a regular file with mode 0600, without symlinks")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) || stat.Nlink != 1 {
		return errors.New("GUI auth file must be owned by the current user and have no hard links")
	}
	if info.Size() > digestCredentialsLimit {
		return errors.New("GUI auth file is too large")
	}
	return nil
}

func loadDigestCredentials(path string) (digestCredentials, error) {
	var credentials digestCredentials
	before, err := digestCredentialInfo(path)
	if err != nil {
		return credentials, err
	}
	// NOFOLLOW and NONBLOCK close the symlink/FIFO race after Lstat. Validate
	// the opened inode too, before consuming any credential bytes.
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return credentials, errors.New("cannot open GUI auth file")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) {
		return credentials, errors.New("GUI auth file changed while opening")
	}
	if err := validateDigestCredentialInfo(opened); err != nil {
		return credentials, err
	}
	data, err := io.ReadAll(io.LimitReader(file, digestCredentialsLimit+1))
	if err != nil || len(data) > digestCredentialsLimit || !utf8.Valid(data) {
		return credentials, errors.New("cannot read bounded GUI auth file")
	}
	invalid := errors.New("invalid GUI auth credential document")
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return digestCredentials{}, invalid
	}
	seen := make(map[string]bool, 4)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return digestCredentials{}, invalid
		}
		name, ok := token.(string)
		if !ok || seen[name] {
			return digestCredentials{}, invalid
		}
		seen[name] = true
		switch name {
		case "version":
			err = decoder.Decode(&credentials.Version)
		case "username":
			err = decoder.Decode(&credentials.Username)
		case "realm":
			err = decoder.Decode(&credentials.Realm)
		case "ha1":
			err = decoder.Decode(&credentials.HA1)
		default:
			return digestCredentials{}, invalid
		}
		if err != nil {
			return digestCredentials{}, invalid
		}
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return digestCredentials{}, invalid
	}
	if decoder.Decode(new(any)) != io.EOF || len(seen) != 4 || credentials.Version != 1 || credentials.Realm != "Ensemble" || !validDigestUsername(credentials.Username) {
		return digestCredentials{}, invalid
	}
	digest, err := hex.DecodeString(credentials.HA1)
	if err != nil || len(digest) != sha256.Size || credentials.HA1 != strings.ToLower(credentials.HA1) {
		return digestCredentials{}, invalid
	}
	return credentials, nil
}

func validDigestUsername(username string) bool {
	if len(username) == 0 || len(username) > 256 {
		return false
	}
	for _, character := range username {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '.' || character == '_' || character == '-') {
			return false
		}
	}
	return true
}
