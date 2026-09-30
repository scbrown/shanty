package segments

import (
	"strings"
	"testing"
)

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

// aegis-apfuey: the REAL `st crew --governor` output on vati, 2026-09-30 18:12Z.
// st moved to one line per lane; the old parser read "base" as an unknown state
// and the bar showed "usage ?" on every refresh while st exited 0.
const stGovernorPerLane = `base ok 34/70/1644 36/70/535644 live 9/9 policy=vati freshest[five_hour@vati,seven_day@vati] no restriction declared
codex ok ?/?/? 100/-/256011 live 0/9 policy=vati freshest[seven_day@vati] dispatch only P0 and above
balance 1.82x — prefer codex (base 3.15x vs codex 1.73x, band 1.5x)
  vati dearing claude live
  macbookair-stiwi hammond claude live`

func TestGovernorPerLaneOutputRendersTheBaseLane(t *testing.T) {
	got := strip(renderGovernor(stGovernorPerLane))
	if got != "Δ 34%5h · 36%7d" {
		t.Fatalf("got %q", got)
	}
	// The codex lane's P0 floor must not leak onto the base (Claude) bar.
	if contains(renderGovernor(stGovernorPerLane), colRed) {
		t.Fatal("base shows no tier, so the bar must not be red")
	}
}

func TestGovernorPerLaneTierIsReadAfterTheRoster(t *testing.T) {
	out := "base ok 53/70/100 26/45/900 live 9/9 policy=vati freshest[five_hour@vati] dispatch only P1 and above\n"
	raw := renderGovernor(out)
	if !contains(raw, colRed) || !contains(strip(raw), "P1+ ONLY") {
		t.Fatalf("an engaged base tier must show red and teach: %q", strip(raw))
	}
}

func TestGovernorPerLaneLostAndUnpublishedStayHonest(t *testing.T) {
	if got := strip(renderGovernor("base lost live 0/9 policy=vati freshest[] no restriction declared\n")); got != "⚠ usage ?" && !contains(got, "usage ?") {
		t.Fatalf("a lost base lane must stay loud, got %q", got)
	}
	got := strip(renderGovernor("base ok ?/?/? 36/70/5 live 9/9 policy=vati freshest[seven_day@vati] no restriction declared\n"))
	if got != "Δ ?5h · 36%7d" {
		t.Fatalf("an unpublished window must read ?5h, never a number: %q", got)
	}
}

// aegis-apfuey step 2: the bar pins `st crew --governor --json` v1. Shaped from
// the real output on vati (2026-09-30), trimmed to the fields the bar reads plus
// neighbours it must ignore.
const stGovernorJSONv1 = `{"version": 1, "scope": "fleet", "host": "vati", "complete": true,
 "governors": {
  "base": {"live": 9, "max_agents": 9, "signal_lost": false, "why": "five_hour 35%; seven_day 36%",
   "windows": {"five_hour": {"published": true, "pct": 35, "next": 70, "reset_seconds": 1228},
               "seven_day": {"published": true, "pct": 36, "next": 70, "reset_seconds": 535228}},
   "effect": "no restriction declared"},
  "codex": {"live": 0, "max_agents": 9, "signal_lost": false,
   "windows": {"five_hour": {"published": false, "pct": null, "next": null, "reset_seconds": null},
               "seven_day": {"published": true, "pct": 100, "next": null, "reset_seconds": 255595}},
   "effect": "dispatch only P0 and above"}},
 "balance": {"prefer": "codex"}}`

func TestGovernorJSONRendersTheBaseLane(t *testing.T) {
	out, ok := renderGovernorJSON(stGovernorJSONv1)
	if !ok || strip(out) != "Δ 35%5h · 36%7d" {
		t.Fatalf("ok=%v got %q", ok, strip(out))
	}
	if contains(out, colRed) {
		t.Fatal("the codex lane's P0 floor must not paint the Claude bar red")
	}
}

func TestGovernorJSONUnknownVersionIsLoudNotBlind(t *testing.T) {
	// The CONTROL sattler asked for: a schema this bar does not know must say so.
	out, ok := renderGovernorJSON(strings.Replace(stGovernorJSONv1, `"version": 1`, `"version": 2`, 1))
	if !ok || !strings.Contains(strip(out), "usage schema v2 unsupported") {
		t.Fatalf("ok=%v got %q", ok, strip(out))
	}
}

func TestGovernorJSONWithoutBarFieldsFallsBackToProse(t *testing.T) {
	// An st from before the windows/effect fields: ok=false, so Render falls back.
	old := `{"version": 1, "governors": {"base": {"live": 9, "signal_lost": false, "readings": {}}}}`
	if _, ok := renderGovernorJSON(old); ok {
		t.Fatal("a v1 answer without windows must fall back, not render")
	}
	if _, ok := renderGovernorJSON("not json"); ok {
		t.Fatal("unparseable output must fall back, not render")
	}
}

func TestGovernorJSONTierLostAndUnpublished(t *testing.T) {
	tier := strings.Replace(stGovernorJSONv1, `"effect": "no restriction declared"`, `"effect": "dispatch only P1 and above"`, 1)
	out, _ := renderGovernorJSON(tier)
	if !contains(out, colRed) || !strings.Contains(strip(out), "P1+ ONLY") {
		t.Fatalf("an engaged base tier must be red and teach: %q", strip(out))
	}
	lost := strings.Replace(stGovernorJSONv1, `"signal_lost": false, "why"`, `"signal_lost": true, "why"`, 1)
	if out, _ := renderGovernorJSON(lost); !strings.Contains(strip(out), "usage ?") {
		t.Fatalf("a lost base signal must stay loud: %q", strip(out))
	}
	unpub := strings.Replace(stGovernorJSONv1, `"five_hour": {"published": true, "pct": 35, "next": 70`, `"five_hour": {"published": false, "pct": null, "next": null`, 1)
	if out, _ := renderGovernorJSON(unpub); strip(out) != "Δ ?5h · 36%7d" {
		t.Fatalf("an unpublished window reads ?5h, never a number: %q", strip(out))
	}
}
