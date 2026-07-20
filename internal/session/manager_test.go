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
