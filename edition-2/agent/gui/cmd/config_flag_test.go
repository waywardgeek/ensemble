package main

import "testing"

func TestConfigPathFromArgs(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
		bad  bool
	}{
		{nil, "", false},
		{[]string{"--vendor", "openai"}, "", false},
		{[]string{"--config", "/tmp/settings.json"}, "/tmp/settings.json", false},
		{[]string{"--config=relative.json", "--port", "8089"}, "relative.json", false},
		{[]string{"-config", "relative.json"}, "relative.json", false},
		{[]string{"--config"}, "", true},
		{[]string{"--config="}, "", true},
		{[]string{"--config", "--port"}, "", true},
		{[]string{"--config", "a", "--config=b"}, "", true},
		{[]string{"--", "--config", "a"}, "", false},
	} {
		got, err := configPathFromArgs(tc.args)
		if (err != nil) != tc.bad || !tc.bad && got != tc.want {
			t.Errorf("%v: got %q, %v", tc.args, got, err)
		}
	}
}
