package gradeedition2

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter8 preserves the original four settings weights and grades inherited
// behavior through its separate GUI/core clients. Snapshot checks do not stand
// in for actual budget execution, entry/load validation or browser behavior.
func Chapter8(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	checks := settingsChecks(gui, guiModel)
	previous, err := Chapter7(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch7-parity", Title: "All corrected Chapter 7 checks pass", Points: 20, Earned: 20, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 8 — Build Your Dream GUI", append(checks, parity), ""), nil
}

func settingsChecks(dir, model string) []grade.Check {
	r := &grade.Ch9Result{}
	bin, cleanup, err := grade.Build(dir)
	if err != nil {
		r.BuildErr = err.Error()
		return grade.Ch9Evaluate(r)[:4]
	}
	defer cleanup()
	return settingsExecutableChecks(bin, filepath.Join(dir, "web", "gui"), model)
}

// Chapter8Executable retains the four settings protocol/persistence checks.
// Core budget and actual browser controls need separate C# integration evidence;
// inherited Go consumer/topology scoring is deliberately excluded.
func Chapter8Executable(bin, static, model string) grade.Report {
	return grade.NewTitledReport("Edition 3 C# — Chapter 8 executable settings checks only (80 points)", settingsExecutableChecks(bin, static, model), "")
}

func settingsExecutableChecks(bin, static, model string) []grade.Check {
	r := &grade.Ch9Result{BuildOK: true}
	problems := driveSettings(r, bin, static, model)
	checks := grade.Ch9Evaluate(r)[:4]
	// Report the checked semantics rather than mislabeling the observed wire
	// when the executable uses the chapter's settings_updated spelling.
	checks[0].Title = "update_settings returns a full acknowledged snapshot"
	for i := range checks {
		for j := range checks[i].Details {
			checks[i].Details[j] = strings.ReplaceAll(checks[i].Details[j], "settings_changed", "settings update")
		}
	}
	for i, messages := range problems {
		for _, message := range messages {
			fail(&checks[i], "%s", message)
		}
	}
	checks[3].Details = append(checks[3].Details, "checks full visible preferences and explicit zero budget across actual restart; effective budget, validation, write failure and browser controls remain source/integration proof")
	return checks
}

func readSettings(peer *guiPeer, kind string) (map[string]any, error) {
	messages, err := peer.until(3*time.Second, func(message map[string]any) bool {
		// The public chapter names the acknowledgement settings_updated.
		// Retain the older Go fixture spelling without requiring a product alias.
		return message["type"] == kind || (kind == "settings_changed" && message["type"] == "settings_updated")
	})
	if err != nil {
		return nil, err
	}
	snapshot := object(messages[len(messages)-1]["settings"])
	if snapshot == nil {
		return nil, fmt.Errorf("%s lacks a settings object", kind)
	}
	return snapshot, nil
}

// Only compare fields owned by this exercise. Extra settings do not violate a
// full snapshot, and the fixture values are not prescribed defaults or bounds.
func settingsDifference(snapshot, want map[string]any) string {
	for _, key := range []string{"theme", "tts_enabled", "tts_speed", "font_size", "max_tool_rounds"} {
		expected, needed := want[key]
		if !needed {
			continue
		}
		got, present := snapshot[key]
		if !present || !reflect.DeepEqual(got, expected) {
			return fmt.Sprintf("%s: want present %v, got %v (present=%t)", key, expected, got, present)
		}
	}
	return ""
}

func driveSettings(r *grade.Ch9Result, bin, static, model string) (problems [4][]string) {
	// These are established valid inputs, not the accepted student's clamp
	// endpoints. Dark/speed2 preserve the original broadcast fixture; final system
	// plus false/speed2/font18 distinguish saved preferences from startup defaults.
	patches := []map[string]any{
		{"theme": "light", "tts_enabled": true, "tts_speed": 1.5, "font_size": float64(18)},
		{"theme": "dark", "tts_speed": float64(2)},
		{"tts_enabled": false, "max_tool_rounds": float64(2)},
		{"max_tool_rounds": float64(0)},
		{"theme": "system"},
	}
	expected := map[string]any{}
	check := func(index int, label string, snapshot map[string]any, err error) {
		if err != nil {
			problems[index] = append(problems[index], label+": "+err.Error())
			return
		}
		if difference := settingsDifference(snapshot, expected); difference != "" {
			problems[index] = append(problems[index], label+": "+difference)
		}
	}
	setupFailure := func(err error) {
		for i := range problems {
			problems[i] = append(problems[i], "settings lifecycle: "+err.Error())
		}
	}
	work, err := os.MkdirTemp("", "edition2-settings-")
	if err != nil {
		setupFailure(err)
		return
	}
	defer os.RemoveAll(work)
	static, err = filepath.Abs(static)
	if err != nil {
		setupFailure(err)
		return
	}
	// Settings do not need inference, but the real executable still receives a
	// valid provider configuration. No runtime method is replaced by this peer.
	provider := fakevendor.New([]fakevendor.Reply{{Text: "Settings test response."}})
	defer provider.Close()
	url, stop, err := startGUI(bin, static, model, provider.URL(), work)
	if err != nil {
		setupFailure(err)
		return
	}
	defer func() {
		if stop != nil {
			stop()
		}
	}()
	err = func() error {
		first, err := openGUIPeer(url)
		if err != nil {
			return err
		}
		defer first.close()
		second, err := openGUIPeer(url)
		if err != nil {
			return err
		}
		defer second.close()
		for _, peer := range []*guiPeer{first, second} {
			if err := peer.send("subscribe", ""); err != nil {
				return err
			}
			initial, err := readSettings(peer, "current_settings")
			if err != nil {
				problems[1] = append(problems[1], "initial subscription: "+err.Error())
				continue
			}
			// Default values are a product choice; presence and JSON types are not.
			_, theme := initial["theme"].(string)
			_, enabled := initial["tts_enabled"].(bool)
			_, speed := initial["tts_speed"].(float64)
			font, hasFont := initial["font_size"].(float64)
			rounds, hasRounds := initial["max_tool_rounds"].(float64)
			if !theme || !enabled || !speed || !hasFont || font != float64(int(font)) || !hasRounds || rounds != float64(int(rounds)) {
				problems[1] = append(problems[1], "initial snapshot lacks five correctly typed visible settings")
			}
		}
		for i, patch := range patches {
			for key, value := range patch {
				expected[key] = value
			}
			if err := first.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": patch}); err != nil {
				return err
			}
			echo, e := readSettings(first, "settings_changed")
			broadcast, b := readSettings(second, "settings_changed")
			check(0, fmt.Sprintf("sender patch %d", i+1), echo, e)
			check(2, fmt.Sprintf("second client patch %d", i+1), broadcast, b)
			// Populate the original evaluator with its original scenarios, while the
			// supplements also examine sparse updates, false and zero on both clients.
			if i == 0 {
				r.RoundtripSent = patch
				r.RoundtripReceived = echo
				r.RoundtripOK = e == nil
				if e != nil {
					r.RoundtripErr = e.Error()
				}
			}
			if i == 1 {
				r.BroadcastReceived = broadcast
				r.BroadcastOK = b == nil
				if b != nil {
					r.BroadcastErr = b.Error()
				}
			}
		}
		fresh, err := openGUIPeer(url)
		if err != nil {
			return err
		}
		defer fresh.close()
		if err := fresh.send("subscribe", ""); err != nil {
			return err
		}
		snapshot, e := readSettings(fresh, "current_settings")
		r.OnConnectReceived = snapshot
		r.OnConnectOK = e == nil
		if e != nil {
			r.OnConnectErr = e.Error()
		}
		check(1, "fresh subscriber", snapshot, e)
		return nil
	}()
	if err != nil {
		setupFailure(err)
		return
	}
	// Close peers and the actual first process before starting a second process
	// in this same directory. A reconnect to the original server proves no disk read.
	stop()
	stop = nil
	url, stop, err = startGUI(bin, static, model, provider.URL(), work)
	if err != nil {
		problems[3] = append(problems[3], "restart: "+err.Error())
		return
	}
	restarted, err := openGUIPeer(url)
	if err != nil {
		problems[3] = append(problems[3], err.Error())
		return
	}
	defer restarted.close()
	if err = restarted.send("subscribe", ""); err != nil {
		problems[3] = append(problems[3], err.Error())
		return
	}
	snapshot, e := readSettings(restarted, "current_settings")
	r.PersistReceived = snapshot
	r.PersistOK = e == nil
	if e != nil {
		r.PersistErr = e.Error()
	}
	check(3, "restarted subscriber", snapshot, e)
	return
}
