package segments

import "testing"

// strip removes tmux colour markup so a test asserts on TEXT, not styling.
func strip(s string) string {
	out := ""
	depth := 0
	for _, r := range s {
		switch {
		case r == '#':
			// lookahead is unnecessary: every marker here is #[...]
		case r == '[':
			depth++
		case r == ']':
			if depth > 0 {
				depth--
			}
		case depth == 0:
			out += string(r)
		}
	}
	return out
}

func TestUsageRendersBothBudgets(t *testing.T) {
	got := strip(renderGovernor("ok 45/50 24/45"))
	if got != "Δ 45%5h · 24%7d" {
		t.Errorf("got %q", got)
	}
}

// THE POINT OF THE WHOLE SEGMENT. A blind governor must never render a number —
// a stale percentage on the bar would silently undo the fail-safe, which is that
// blindness alarms every pass.
func TestBlindNeverRendersANumber(t *testing.T) {
	for _, in := range []string{"lost", "", "garbage", "ok 45"} {
		got := strip(renderGovernor(in))
		for _, r := range got {
			if r >= '0' && r <= '9' {
				t.Fatalf("input %q rendered a digit: %q", in, got)
			}
		}
		if got == "" {
			t.Errorf("input %q rendered blank; blank reads as 'nothing to report'", in)
		}
	}
}

// `off` is the one silent case, and deliberately so: shanty is usable without a
// governor and must not nag about a feature nobody turned on.
func TestOffIsSilentNotLoud(t *testing.T) {
	if got := renderGovernor("off"); got != hidden {
		t.Errorf("off rendered %q, want the silent empty", got)
	}
}

func TestEngagedTierIsNamedAndRed(t *testing.T) {
	raw := renderGovernor("ok 53/70 26/45 dispatch only P1 and above [five_hour >= 50%]")
	got := strip(raw)
	if want := "Δ 53%5h · 26%7d P1+ ONLY"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if !contains(raw, colRed) {
		t.Error("an engaged tier must render red")
	}
}

func TestColourTracksTheNEARESTTierNotTheRawNumber(t *testing.T) {
	// 44 is 6 from the five_hour tier (amber) but the SAME number is already past
	// the first seven_day tier elsewhere — this is why the raw number cannot drive
	// the colour and st sends the next threshold.
	if raw := renderGovernor("ok 44/50 20/45"); !contains(raw, colOrange) {
		t.Error("44/50 is within the approach band and must be amber")
	}
	if raw := renderGovernor("ok 20/50 20/45"); !contains(raw, colGreen) {
		t.Error("20/50 is far from any tier and must be green")
	}
	// Approaching on the SEVEN-DAY window alone is still approaching. A segment
	// that only watched five_hour would call this green while the weekly budget —
	// the one that takes DAYS to refill — was two points from engaging.
	if raw := renderGovernor("ok 10/50 43/45"); !contains(raw, colOrange) {
		t.Error("seven_day approaching must colour the pair amber")
	}
}

func TestUnreadableWindowIsAQuestionNotZero(t *testing.T) {
	got := strip(renderGovernor("ok 45/50 ?/?"))
	if got != "Δ 45%5h · ?7d" {
		t.Errorf("got %q; an unread window must not render as a number", got)
	}
}

func TestNoHigherTierIsNotApproaching(t *testing.T) {
	// 97 with no tier above it: nothing to approach. It should not be amber for
	// lack of a threshold — it is red only because a tier is engaged.
	if raw := renderGovernor("ok 97/- 20/45"); contains(raw, colOrange) {
		t.Error("a window above every tier has nothing to approach")
	}
}

func TestShortTierTeachesWithoutOverflowing(t *testing.T) {
	cases := map[string]string{
		"dispatch only P0 and above [five_hour >= 70%]":                  "P0+ ONLY",
		"dispatch only P1 and above [five_hour >= 50%]":                  "P1+ ONLY",
		"only support crew runs [five_hour >= 80%]":                      "SUPPORT ONLY",
		"FULL STOP — every agent pushes its work, then stops [x >= 95%]": "DRAIN",
	}
	for in, want := range cases {
		if got := shortTier(in); got != want {
			t.Errorf("shortTier(%q) = %q, want %q", in, got, want)
		}
	}
	// An unrecognised shape degrades to a visible truncation, never to silence:
	// a tier nobody anticipated must still be obvious on the bar.
	got := shortTier("some future restriction nobody wrote yet [x >= 60%]")
	if got == "" {
		t.Error("an unknown tier shape must not vanish")
	}
	if len(got) > 20 {
		t.Errorf("unknown tier shape not truncated: %q", got)
	}
}

func TestUsageIsRegisteredAndFleetWide(t *testing.T) {
	if _, ok := Registry["usage"]; !ok {
		t.Fatal("usage is not registered")
	}
	if Registry["usage"].Name() != "usage" {
		t.Error("segment name mismatch")
	}
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
