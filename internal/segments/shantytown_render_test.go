package segments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scbrown/shanty/internal/stread"
)

// fakeST installs a stub `st` that answers from the given canned replies, keyed by
// the joined argument list. Anything unlisted exits nonzero, so a segment reaching
// for a call the test did not anticipate shows up as a failure instead of quietly
// passing.
//
// The renderings below are the acceptance criteria for this bar: a busy agent, an
// idle agent, and an agent whose plate cannot be read must produce three DISTINCT
// outputs and none of them may be blank. Two of the three can be observed on a
// live fleet; a genuinely idle agent with an empty plate cannot be conjured on
// demand, which is why all three are pinned here.
func fakeST(t *testing.T, replies map[string]string) {
	t.Helper()
	dir := t.TempDir()
	var sb strings.Builder
	sb.WriteString("#!/bin/sh\ncase \"$*\" in\n")
	for args, out := range replies {
		fmt.Fprintf(&sb, "  %q) cat <<'REPLY'\n%s\nREPLY\n  ;;\n", args, out)
	}
	sb.WriteString("  *) echo \"unexpected st call: $*\" >&2; exit 9 ;;\nesac\n")

	bin := filepath.Join(dir, "st")
	if err := os.WriteFile(bin, []byte(sb.String()), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHANTY_ST_BIN", bin)
	t.Setenv("SHANTY_SEG_NOCACHE", "1")
	t.Setenv("SHANTY_AGENT", "villiers")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // never touch the real mark registry
	stread.ResetBinForTest()
	t.Cleanup(stread.ResetBinForTest)
}

// plain strips tmux colour codes so assertions read as what an operator sees.
func plain(s string) string {
	var out strings.Builder
	for {
		i := strings.Index(s, "#[")
		if i < 0 {
			out.WriteString(s)
			return out.String()
		}
		out.WriteString(s[:i])
		j := strings.Index(s[i:], "]")
		if j < 0 {
			return out.String()
		}
		s = s[i+j+1:]
	}
}

const busyCrew = "  villiers    worker    up   current  busy    st-villiers"

func TestTaskRendersAHeldItem(t *testing.T) {
	fakeST(t, map[string]string{
		"anchor villiers --short": "ss-5x22f",
		"anchor villiers": "  You are villiers — worker, reports to moneypenny.\n" +
			"\n  ON YOUR PLATE\n" +
			"    ▶ ss-5x22f  status display: missing stats and crew name        (in_progress)\n",
		"crew": busyCrew,
	})
	got := plain(Task{}.Render())
	if !strings.Contains(got, "ss-5x22f") {
		t.Errorf("render %q is missing the item id", got)
	}
	if !strings.Contains(got, "status display") {
		t.Errorf("render %q is missing the task summary", got)
	}
	if strings.Contains(got, "⚠") {
		t.Errorf("render %q warns about a plate it read fine", got)
	}
}

func TestTaskRendersAnIdleAgentWithAnEmptyPlate(t *testing.T) {
	// st legitimately names no item and calls the agent idle. That is an ANSWER, so
	// it gets words — an empty segment here would be indistinguishable from the bar
	// being broken.
	fakeST(t, map[string]string{
		"anchor villiers --short": "",
		"crew":                    "  villiers    worker    up   current  idle    st-villiers",
	})
	got := plain(Task{}.Render())
	if got == "" {
		t.Fatal("an idle agent rendered a blank segment")
	}
	if strings.Contains(got, "⚠") {
		t.Errorf("render %q warns about a plate that is simply empty", got)
	}
	if !strings.Contains(got, "nothing held") {
		t.Errorf("render %q does not say the plate is empty", got)
	}
}

func TestTaskIsLoudWhenABusyAgentsPlateCannotBeRead(t *testing.T) {
	// THE bug this bar was rebuilt for. `st anchor --short` answers a lookup it
	// cannot resolve with empty output and a ZERO exit, so a naive bar renders
	// blank — consistently, silently, and looking exactly like an idle agent.
	// Cross-checking st's own verdict turns it into a visible contradiction.
	fakeST(t, map[string]string{
		"anchor villiers --short": "",
		"crew":                    busyCrew,
	})
	got := plain(Task{}.Render())
	if !strings.Contains(got, "⚠") {
		t.Errorf("render %q must warn: st calls this agent busy but names no item", got)
	}
	if strings.Contains(got, "nothing held") {
		t.Errorf("render %q claims an empty plate for a busy agent", got)
	}
}

func TestTheThreePlateRenderingsAreDistinct(t *testing.T) {
	// The acceptance criterion, stated as a test: three states, three renderings,
	// no blanks.
	renders := map[string]string{}
	t.Run("held", func(t *testing.T) {
		fakeST(t, map[string]string{
			"anchor villiers --short": "ss-1",
			"anchor villiers":         "  ON YOUR PLATE\n    ▶ ss-1  a real item        (open)\n",
			"crew":                    busyCrew,
		})
		renders["held"] = plain(Task{}.Render())
	})
	t.Run("idle", func(t *testing.T) {
		fakeST(t, map[string]string{
			"anchor villiers --short": "",
			"crew":                    "  villiers    worker    up   current  idle    st-villiers",
		})
		renders["idle"] = plain(Task{}.Render())
	})
	t.Run("unreadable", func(t *testing.T) {
		fakeST(t, map[string]string{
			"anchor villiers --short": "",
			"crew":                    busyCrew,
		})
		renders["unreadable"] = plain(Task{}.Render())
	})

	seen := map[string]string{}
	for state, out := range renders {
		if strings.TrimSpace(out) == "" {
			t.Errorf("state %q rendered blank", state)
		}
		if other, dup := seen[out]; dup {
			t.Errorf("states %q and %q render identically as %q", other, state, out)
		}
		seen[out] = state
	}
}

func TestCrewIDShowsMarkNameRoleAndState(t *testing.T) {
	fakeST(t, map[string]string{"crew": busyCrew})
	got := plain(CrewID{}.Render())
	for _, want := range []string{"villiers", "wkr", "busy"} {
		if !strings.Contains(got, want) {
			t.Errorf("render %q is missing %q", got, want)
		}
	}
	// A mark is assigned on first sight, so the pane is identifiable immediately.
	if strings.HasPrefix(got, "villiers") {
		t.Errorf("render %q carries no mark", got)
	}
}

func TestCrewIDFlagsStaleSettings(t *testing.T) {
	// st reports the agent is running settings older than the file on disk. The
	// pane looks healthy and its hooks are whatever the file said at launch, so
	// this is exactly the kind of thing a bar must not swallow.
	fakeST(t, map[string]string{
		"crew": "  villiers    worker    up   STALE  busy    st-villiers",
	})
	got := plain(CrewID{}.Render())
	if !strings.Contains(got, "STALE") {
		t.Errorf("render %q hides that the agent's settings are stale", got)
	}
}

func TestCrewIDKeepsIdentityWhenStateIsUnknowable(t *testing.T) {
	// st answered, but not about us. The identity half is still true and still
	// useful; dropping the whole segment would throw it away.
	fakeST(t, map[string]string{
		"crew": "  bond    worker    up   current  busy    st-bond",
	})
	got := plain(CrewID{}.Render())
	if !strings.Contains(got, "villiers") {
		t.Errorf("render %q lost the identity it knew", got)
	}
	if !strings.Contains(got, "⚠") {
		t.Errorf("render %q does not flag the missing crew state", got)
	}
}

func TestStatsRendersTheNumbers(t *testing.T) {
	fakeST(t, map[string]string{
		"stats villiers": "st stats — last 24h\n" +
			"  villiers    events=412  files=17  stops=6  tokens_in=180000 tokens_out=42000",
	})
	got := plain(Stats{}.Render())
	for _, want := range []string{"412", "17f", "222ktok"} {
		if !strings.Contains(got, want) {
			t.Errorf("render %q is missing %q", got, want)
		}
	}
}

func TestStatsSaysOffRatherThanZeroWhenNobodyIsCounting(t *testing.T) {
	// `st stats` is wired on every deployment but its numbers come from a local
	// capture store that only exists once the harness hooks feeding it are
	// installed. Rendering zeros would claim a crew did nothing when in fact
	// nobody was counting — and it is the live state of this deployment today.
	dir := t.TempDir()
	bin := filepath.Join(dir, "st")
	script := "#!/bin/sh\n" +
		"echo 'st stats — no capture store yet (.shanty/stats.sqlite absent).'\nexit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHANTY_ST_BIN", bin)
	t.Setenv("SHANTY_SEG_NOCACHE", "1")
	t.Setenv("SHANTY_AGENT", "villiers")
	stread.ResetBinForTest()
	t.Cleanup(stread.ResetBinForTest)

	got := plain(Stats{}.Render())
	if !strings.Contains(got, "off") {
		t.Errorf("render %q should say the counting is off, not show numbers", got)
	}
	if strings.Contains(got, "0⚡") {
		t.Errorf("render %q reports zero activity when nothing is being measured", got)
	}
}

func TestReadsAreNonMarking(t *testing.T) {
	// The bar polls every few seconds and must never consume state. st documents
	// `inbox --count` and `anchor --events` as reads that mark nothing; the bar has
	// to use exactly those, never the listing or draining forms.
	calls := filepath.Join(t.TempDir(), "calls")
	dir := t.TempDir()
	bin := filepath.Join(dir, "st")
	script := "#!/bin/sh\necho \"$*\" >> " + calls + "\necho 0\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHANTY_ST_BIN", bin)
	t.Setenv("SHANTY_SEG_NOCACHE", "1")
	t.Setenv("SHANTY_AGENT", "villiers")
	stread.ResetBinForTest()
	t.Cleanup(stread.ResetBinForTest)

	Events{}.Render()
	Inbox{}.Render()

	b, err := os.ReadFile(calls)
	if err != nil {
		t.Fatalf("reading call log: %v", err)
	}
	log := string(b)
	if !strings.Contains(log, "anchor villiers --events") {
		t.Errorf("events segment did not use the non-marking --events read: %q", log)
	}
	if !strings.Contains(log, "inbox --count villiers") {
		t.Errorf("inbox segment did not use the non-marking --count read: %q", log)
	}
	for _, forbidden := range []string{"--read", "drain"} {
		if strings.Contains(log, forbidden) {
			t.Errorf("the bar called a state-consuming form %q: %q", forbidden, log)
		}
	}
}

func TestAdministratorHoldingNothingIsNotAFault(t *testing.T) {
	// The administrator role exists to stay free to coordinate and is never assigned
	// implementation work, so its empty plate is the EXPECTED state. Warning about it
	// produces a red that is known-false every time it is drawn, and a status surface
	// carrying one permanent false red teaches its reader to discount every red on it.
	fakeST(t, map[string]string{
		"anchor moneypenny --short": "",
		"crew": "  moneypenny  administrator  up  current  busy  st-moneypenny\n" +
			"  bond        worker         up  current  idle  st-bond\n" +
			"  q           worker         up  current  busy  st-q",
	})
	t.Setenv("SHANTY_AGENT", "moneypenny")

	got := plain(Task{}.Render())
	if strings.Contains(got, "⚠") {
		t.Errorf("render %q warns about an administrator's empty plate", got)
	}
	if strings.Contains(got, "no item") {
		t.Errorf("render %q reports an absence for a role that never holds one: %q", got, got)
	}
	// It must say something TRUE instead of nothing — the slot is not wasted.
	if strings.TrimSpace(got) == "" {
		t.Error("administrator render is blank; it should show dispatch state")
	}
	if !strings.Contains(got, "free") {
		t.Errorf("render %q does not report the dispatch state (1 agent is idle)", got)
	}
}

func TestAdministratorRuleIsKeyedOnRoleNotName(t *testing.T) {
	// A future administrator must inherit the treatment, and the current one must not
	// keep it if the role moves. Same agent NAME, different role, opposite rendering.
	for _, tc := range []struct {
		role      string
		wantWarn  bool
		situation string
	}{
		{"administrator", false, "coordinating role: empty plate is expected"},
		{"worker", true, "worker with no item while busy: real signal, stay loud"},
	} {
		fakeST(t, map[string]string{
			"anchor moneypenny --short": "",
			"crew":                      "  moneypenny  " + tc.role + "  up  current  busy  st-moneypenny",
		})
		t.Setenv("SHANTY_AGENT", "moneypenny")
		got := plain(Task{}.Render())
		if warned := strings.Contains(got, "⚠"); warned != tc.wantWarn {
			t.Errorf("role %q (%s): warned=%v, want %v — render was %q",
				tc.role, tc.situation, warned, tc.wantWarn, got)
		}
	}
}

func TestAdministratorSummaryCountsIdleWithLiveWorkSeparately(t *testing.T) {
	// An agent st calls idle whose background work is still running is NOT free, and
	// counting it as free is the single number that most often makes a dispatch
	// decision wrong.
	fakeST(t, map[string]string{
		"anchor moneypenny --short": "",
		"crew": "  moneypenny  administrator  up  current  busy      st-moneypenny\n" +
			"  bond        worker         up  current  idle      st-bond\n" +
			"  q           worker         up  current  idle+1sh  st-q\n" +
			"  r           lead           up  current  waiting   st-r",
	})
	t.Setenv("SHANTY_AGENT", "moneypenny")
	got := plain(Task{}.Render())
	if !strings.Contains(got, "1 free") {
		t.Errorf("render %q: only bond is genuinely free", got)
	}
	if !strings.Contains(got, "1 idle+live") {
		t.Errorf("render %q: q is idle with a live shell and must not count as free", got)
	}
	if !strings.Contains(got, "1 need eyes") {
		t.Errorf("render %q: r is waiting and needs a human", got)
	}
}

func TestCrewIDShowsWhyItThinksSo(t *testing.T) {
	// The coordinator's ask: expose WHY the bar believes an agent is busy or idle, so
	// the verdict can be weighed against a competing idle signal rather than just
	// believed or ignored.
	fakeST(t, map[string]string{"crew": "  villiers  worker  up  current  busy+2sh  st-villiers"})
	got := plain(CrewID{}.Render())
	if !strings.Contains(got, "2 shells live") {
		t.Errorf("render %q does not explain the verdict", got)
	}
}

func TestCrewIDDoesNotPaintIdleWithLiveWorkAsCalm(t *testing.T) {
	fakeST(t, map[string]string{"crew": "  villiers  worker  up  current  idle+1sh  st-villiers"})
	got := CrewID{}.Render()
	if strings.Contains(got, colGreen) {
		t.Errorf("idle-with-live-work rendered in the calm green of a free agent: %q", got)
	}
	if !strings.Contains(plain(got), "1 shell live") {
		t.Errorf("render %q hides that work is still running", plain(got))
	}
}
