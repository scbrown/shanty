package stread

import (
	"strconv"
	"strings"
)

// Stat is one agent's slice of `st stats` — what that crew member actually did in
// the window st reports on.
//
// Captured is the field that keeps this honest. `st stats` is wired as a command
// on every deployment, but its numbers come from a LOCAL capture store that only
// exists once the harness hooks that feed it are installed. Absent hooks means
// absent store, and st says so plainly. Rendering that as zeros would claim a
// crew did nothing when the truth is that nobody was counting.
type Stat struct {
	Captured  bool // false = st has no capture store; the numbers below are meaningless
	Events    int  // tool calls plus stops seen in the window
	Files     int  // distinct files touched
	Stops     int  // harness stops
	TokensIn  int
	TokensOut int
}

// Tokens is the total token traffic attributed to the agent.
func (s Stat) Tokens() int { return s.TokensIn + s.TokensOut }

// noStoreMarker is st's own wording when the capture store does not exist. st
// pairs it with a nonzero exit, so matching the text is what tells "nobody is
// counting" apart from "st broke".
const noStoreMarker = "no capture store"

// Stats reads one agent's numbers.
//
// A Stat with Captured false and a nil error is the "nothing is being counted"
// answer — a real, renderable state, not an error. An error means we could not
// ask st at all.
func Stats(agent string) (Stat, error) {
	out, err := Run("stats", agent)
	if strings.Contains(out, noStoreMarker) {
		return Stat{}, nil
	}
	if err != nil {
		return Stat{}, err
	}
	return parseStats(out, agent), nil
}

// parseStats reads the agent's row out of st's report. The row is
// `<agent> events=N files=N stops=N tokens_in=N tokens_out=N`, so the parse keys
// on the NAME and then on key=value pairs — no column positions, and unknown keys
// are ignored rather than shifting anything.
func parseStats(out, agent string) Stat {
	s := Stat{Captured: true}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != agent {
			continue
		}
		for _, f := range fields[1:] {
			k, v, ok := strings.Cut(f, "=")
			if !ok {
				continue
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			switch k {
			case "events":
				s.Events = n
			case "files":
				s.Files = n
			case "stops":
				s.Stops = n
			case "tokens_in":
				s.TokensIn = n
			case "tokens_out":
				s.TokensOut = n
			}
		}
		return s
	}
	// st reported a store but no row for this agent: the window holds no activity
	// for them. Captured stays true — zeros here are a measurement, not a gap.
	return s
}
