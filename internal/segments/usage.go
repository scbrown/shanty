package segments

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/scbrown/shanty/internal/stread"
)

// Usage renders BOTH Claude usage budgets and the throttle tier in force.
//
// WHY IT EXISTS. The number moves fast and nothing an operator watches said so:
// five_hour went 11% -> 41% -> 49% -> 53% in a single evening, crossing its first
// tier, and the only ways to see it were a full `st tend -n` supervise pass or a
// hand-rolled authenticated PromQL query. By the time anyone thought to check,
// the crew could already be shedding work.
//
// FLEET-WIDE, not per-agent: the budget belongs to the account, not the pane, so
// this segment takes no #{session_name} and must NOT be added to perAgentSegments.
// Same class as `crew`.
//
// It holds NO CREDENTIAL. Prometheus requires basic auth, but st already reads it
// and `st crew --governor` prints the verdict, so the bar shells out exactly like
// every other st-backed segment here. A status segment holding a credential is a
// status segment that leaks one — into a tmux config, a process list, and every
// screenshot of the bar.
type Usage struct{}

func (u Usage) Name() string { return "usage" }

func (u Usage) Render() string {
	if !stread.Installed() {
		return hidden
	}
	out, err := stread.Run("crew", "--governor")
	if err != nil {
		// A governor we cannot ask about is not a governor reading zero.
		// stread.Run caches only SUCCESSFUL answers, so this stays live rather
		// than pinning a transient failure in place for the whole TTL.
		return loud("usage ?")
	}
	return renderGovernor(out)
}

// governorLine is one parsed `st crew --governor` answer.
type governorLine struct {
	ok        bool // false => blind (lost/off/unparseable); the pcts are meaningless
	off       bool // no governor configured at all
	fiveNow   int
	fiveNext  int // -1 = no higher tier
	sevenNow  int
	sevenNext int
	fiveOK    bool // the window reported a real number (st prints ?/? when not)
	sevenOK   bool
	tier      string // the engaged tier's label, "" when none
}

// parseGovernor reads st's contract:
//
//	ok 45/50 24/45                                       no tier engaged
//	ok 70/80 24/45 dispatch only P0 and above [...]      a tier is in force
//	lost                                                 the signal is unreadable
//	off                                                  no governor configured
//
// The blind answers carry NO DIGITS by construction on st's side, so there is no
// path here that turns one into a percentage — which is the point. A bar showing
// a stale number while the governor is blind would silently undo the governor's
// whole fail-safe, which is that blindness ALARMS every pass.
func parseGovernor(s string) governorLine {
	f := strings.Fields(strings.TrimSpace(s))
	if len(f) == 0 {
		return governorLine{}
	}
	switch f[0] {
	case "off":
		return governorLine{off: true}
	case "ok":
		// fall through
	default: // "lost", or anything we do not recognise
		return governorLine{}
	}
	if len(f) < 3 {
		return governorLine{}
	}
	five, fiveNext, fiveOK := splitBudget(f[1])
	seven, sevenNext, sevenOK := splitBudget(f[2])
	return governorLine{
		ok:      true,
		fiveNow: five, fiveNext: fiveNext, fiveOK: fiveOK,
		sevenNow: seven, sevenNext: sevenNext, sevenOK: sevenOK,
		tier: strings.Join(f[3:], " "),
	}
}

// splitBudget reads `now/next`; next is -1 when st printed `-` (no higher tier),
// and ok is false for `?/?` (the producer does not publish that window).
func splitBudget(s string) (now, next int, ok bool) {
	a, b, found := strings.Cut(s, "/")
	if !found {
		return 0, -1, false
	}
	n, err := strconv.Atoi(a)
	if err != nil {
		return 0, -1, false
	}
	nx := -1
	if b != "-" {
		if v, err := strconv.Atoi(b); err == nil {
			nx = v
		}
	}
	return n, nx, true
}

// approachingBy is how many points below its next tier a budget starts reading as
// amber. Small on purpose: the five-hour number moved 30+ points in an evening, so
// a wide band would sit amber permanently and stop meaning anything.
const approachingBy = 8

// renderGovernor is the pure display half, so every state is testable without st.
func renderGovernor(out string) string {
	g := parseGovernor(out)
	switch {
	case g.off:
		// shanty is usable without a governor; do not nag about a feature the
		// operator never turned on. This is the same judgement as `hidden`.
		return hidden
	case !g.ok:
		return loud("usage ?")
	}

	colour := colGreen
	if nearTier(g.fiveNow, g.fiveNext, g.fiveOK) || nearTier(g.sevenNow, g.sevenNext, g.sevenOK) {
		colour = colOrange
	}
	if g.tier != "" {
		colour = colRed
	}

	// BOTH numbers take the same colour: the tiers are cumulative and the UNION
	// across windows engages, so a single verdict governs the pair. Colouring
	// them separately would imply they can be in different states, which is
	// exactly the misreading the two-budget display exists to prevent.
	body := paint(colour, budgetBare(g.fiveNow, g.fiveOK, "5h")) +
		paint(colDim, " · ") +
		paint(colour, budgetBare(g.sevenNow, g.sevenOK, "7d"))
	if g.tier != "" {
		body += " " + paint(colRed, shortTier(g.tier))
	}
	return paint(colDim, "Δ ") + body
}

// nearTier: is this budget within approachingBy of engaging its next tier? A
// window with no next tier cannot approach one, and a window we could not read
// is not "fine" — it simply does not vote.
func nearTier(now, next int, ok bool) bool {
	return ok && next >= 0 && next-now <= approachingBy
}

// budgetBare renders one window. A window st could not read shows `?5h`, never a
// number — the absent case must not be able to look like a low one.
func budgetBare(pct int, ok bool, label string) string {
	if !ok {
		return "?" + label
	}
	return fmt.Sprintf("%d%%%s", pct, label)
}

var (
	reDispatchFloor = regexp.MustCompile(`dispatch only P(\d+) and above`)
	reTraitsOnly    = regexp.MustCompile(`only ([\w/]+) crew runs`)
)

// shortTier compresses st's full refusal sentence to something that fits a bar
// while still TEACHING — "P1+ ONLY" says what is in force; "53%" says nothing
// about what changed.
//
// It degrades to a truncation rather than to silence or to a guess: an unmatched
// label still shows its leading clause, so a tier shape nobody anticipated is
// visible and obviously unabbreviated instead of vanishing.
func shortTier(label string) string {
	if strings.Contains(label, "FULL STOP") {
		return "DRAIN"
	}
	if m := reDispatchFloor.FindStringSubmatch(label); m != nil {
		return "P" + m[1] + "+ ONLY"
	}
	if m := reTraitsOnly.FindStringSubmatch(label); m != nil {
		return strings.ToUpper(m[1]) + " ONLY"
	}
	head, _, _ := strings.Cut(label, " [")
	if len(head) > 16 {
		head = head[:15] + "…"
	}
	return strings.ToUpper(head)
}
