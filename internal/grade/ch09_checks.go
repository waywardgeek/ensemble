package grade

// Chapter 9 checks — 100 points.
//
// Five checks covering settings management over WebSocket:
//   settings-roundtrip    25  Send update_settings, get settings_changed back with matching values.
//   settings-on-connect   20  New connection receives current_settings on subscribe.
//   settings-broadcast    20  Second client receives settings_changed when first updates.
//   settings-persist      15  Settings survive server restart.
//   ch8-parity            20  All ch8 checks still pass.

import "fmt"

// Ch9Evaluate scores a ch9 submission.
func Ch9Evaluate(r *Ch9Result) []Check {
	return []Check{
		ch9SettingsRoundtrip(r),
		ch9SettingsOnConnect(r),
		ch9SettingsBroadcast(r),
		ch9SettingsPersist(r),
		ch9Ch8Parity(r),
	}
}

func ch9Ready(c *Check, r *Ch9Result) bool {
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return false
	}
	return true
}

// settings-roundtrip: send update_settings, get settings_changed back.
func ch9SettingsRoundtrip(r *Ch9Result) Check {
	c := Check{ID: "settings-roundtrip", Title: "update_settings returns settings_changed", Points: 25, Passed: true, Earned: 25}
	if !ch9Ready(&c, r) {
		return c
	}
	if r.RoundtripErr != "" {
		c.failf("roundtrip error: %s", r.RoundtripErr)
		return c
	}
	if !r.RoundtripOK {
		c.failf("no settings_changed received after update_settings")
		return c
	}

	// Verify sent values appear in the response.
	for k, v := range r.RoundtripSent {
		got, ok := r.RoundtripReceived[k]
		if !ok {
			c.failf("settings_changed missing key %q (sent %v)", k, v)
			return c
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", v) {
			c.failf("settings_changed key %q: sent %v, got %v", k, v, got)
			return c
		}
	}

	c.notef("sent %v, received matching settings_changed", r.RoundtripSent)
	return c
}

// settings-on-connect: new connection receives current_settings on subscribe.
func ch9SettingsOnConnect(r *Ch9Result) Check {
	c := Check{ID: "settings-on-connect", Title: "new connection receives current_settings", Points: 20, Passed: true, Earned: 20}
	if !ch9Ready(&c, r) {
		return c
	}
	if r.OnConnectErr != "" {
		c.failf("on-connect error: %s", r.OnConnectErr)
		return c
	}
	if !r.OnConnectOK {
		c.failf("new connection did not receive current_settings on subscribe")
		return c
	}
	if len(r.OnConnectReceived) == 0 {
		c.failf("current_settings received but empty")
		return c
	}

	c.notef("received current_settings with %d keys on subscribe", len(r.OnConnectReceived))
	return c
}

// settings-broadcast: second client receives settings_changed when first updates.
func ch9SettingsBroadcast(r *Ch9Result) Check {
	c := Check{ID: "settings-broadcast", Title: "settings broadcast to all clients", Points: 20, Passed: true, Earned: 20}
	if !ch9Ready(&c, r) {
		return c
	}
	if r.BroadcastErr != "" {
		c.failf("broadcast error: %s", r.BroadcastErr)
		return c
	}
	if !r.BroadcastOK {
		c.failf("second client did not receive settings_changed")
		return c
	}

	// Verify theme propagated.
	if v, ok := r.BroadcastReceived["theme"]; !ok || v != "dark" {
		c.failf("broadcast settings_changed missing or wrong theme: %v", r.BroadcastReceived)
		return c
	}

	c.notef("second client received settings_changed with correct theme")
	return c
}

// settings-persist: settings survive server restart.
func ch9SettingsPersist(r *Ch9Result) Check {
	c := Check{ID: "settings-persist", Title: "settings survive server restart", Points: 15, Passed: true, Earned: 15}
	if !ch9Ready(&c, r) {
		return c
	}
	if r.PersistErr != "" {
		c.failf("persist error: %s", r.PersistErr)
		return c
	}
	if !r.PersistOK {
		c.failf("settings did not survive server restart")
		return c
	}

	c.notef("theme survived server restart")
	return c
}

// ch8-parity: all ch8 checks still pass.
func ch9Ch8Parity(r *Ch9Result) Check {
	c := Check{ID: "ch8-parity", Title: "chapter 8 behavior is unchanged", Points: 20, Passed: true, Earned: 20}
	if r.Ch8Err != "" {
		c.failf("ch8 harness did not run: %s", r.Ch8Err)
		return c
	}
	if r.Ch8Result == nil {
		c.failf("ch8 harness produced no result")
		return c
	}
	var failed []string
	for _, sub := range Ch8Evaluate(r.Ch8Result) {
		if !sub.Passed {
			failed = append(failed, fmt.Sprintf("%s (%s)", sub.ID, firstOr(sub.Details, "no detail")))
		}
	}
	if len(failed) > 0 {
		c.failf("%d ch8 check(s) now fail: %v", len(failed), failed)
		return c
	}
	c.notef("all ch8 checks still pass")
	return c
}
