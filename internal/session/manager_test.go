package session

import "testing"

func TestFullName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"main", "shanty-main"},
		{"dev", "shanty-dev"},
		{"shanty-main", "shanty-main"}, // already prefixed
		{"shanty-dev", "shanty-dev"},
		{"", "shanty-"},
	}
	for _, tt := range tests {
		got := fullName(tt.input)
		if got != tt.want {
			t.Errorf("fullName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"shanty-main", "main"},
		{"shanty-dev", "dev"},
		{"other-session", "other-session"}, // not a shanty session
		{"shanty-", ""},
	}
	for _, tt := range tests {
		got := displayName(tt.input)
		if got != tt.want {
			t.Errorf("displayName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// The bar needs BOTH halves of st's configuration in the tmux SERVER's environment,
// and they are different directories: SHANTY_ROOT is where st finds who the crew ARE,
// SHANTY_ST_CWD decides which tracker store it reads. Propagating only the second
// yielded "no such agent" for agents that plainly exist.
func TestSegmentEnvCarriesBothHalvesOfStsConfiguration(t *testing.T) {
	for _, want := range []string{"SHANTY_ROOT", "SHANTY_ST_CWD", "SHANTY_ST_BIN"} {
		found := false
		for _, got := range segmentEnv {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("apply does not carry %q into the tmux server: %v", want, segmentEnv)
		}
	}
}
