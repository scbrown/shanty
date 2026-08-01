package crewid

import (
	"os"
	"testing"
)

// isolate points the registry at a scratch config dir so tests never read or
// write the operator's real marks.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	loadOnce = onceReset()
	loaded = nil
}

func TestAssignIsStableWhenTheRosterGrows(t *testing.T) {
	// THE promise of this package. A mark that moves when a crew member joins is
	// worse than no mark: the operator has learned the wall of panes by shape, and
	// re-badging silently invalidates that. This is also why the assignment is
	// stored rather than hashed — a hash is stable per name but its collision
	// handling is not stable across roster changes.
	isolate(t)

	first, err := Assign([]string{"bond", "q", "vesper"})
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	for _, a := range []string{"bond", "q", "vesper"} {
		if first[a] == "" {
			t.Fatalf("no mark assigned for %q: %v", a, first)
		}
	}

	// mallory joins, and the roster now arrives in a different order.
	second, err := Assign([]string{"vesper", "mallory", "bond", "q"})
	if err != nil {
		t.Fatalf("Assign (grown roster): %v", err)
	}
	for _, a := range []string{"bond", "q", "vesper"} {
		if second[a] != first[a] {
			t.Errorf("mark for %q moved from %q to %q when the roster grew",
				a, first[a], second[a])
		}
	}
	if second["mallory"] == "" {
		t.Error("the new crew member got no mark")
	}
	if second["mallory"] == first["bond"] {
		t.Error("the new crew member reused an existing mark")
	}
}

func TestAssignIsDeterministicFromTheRoster(t *testing.T) {
	// Same roster, empty registry, twice: identical result. Assignment walks the
	// SORTED roster against the palette in order, so it cannot depend on which pane
	// happened to render first.
	isolate(t)
	a, err := Assign([]string{"tanner", "felix", "r"})
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	got := map[string]string{}
	for k, v := range a {
		got[k] = v
	}

	isolate(t)
	b, err := Assign([]string{"r", "tanner", "felix"})
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	for k, v := range got {
		if b[k] != v {
			t.Errorf("agent %q got %q on one run and %q on another", k, v, b[k])
		}
	}
}

func TestMarksAreDistinct(t *testing.T) {
	isolate(t)
	roster := []string{
		"bond", "felix", "goodnight", "leiter", "mathis", "moneypenny",
		"pepper", "q", "r", "tanner", "vesper", "villiers", "mallory",
	}
	m, err := Assign(roster)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	seen := map[string]string{}
	for _, a := range roster {
		mark := m[a]
		if mark == "" {
			t.Errorf("%q got no mark", a)
			continue
		}
		if other, dup := seen[mark]; dup {
			t.Errorf("%q and %q share the mark %q — two panes would look alike",
				other, a, mark)
		}
		seen[mark] = a
	}
}

func TestAssignIsIdempotent(t *testing.T) {
	isolate(t)
	if _, err := Assign([]string{"bond"}); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	p, _ := Path()
	before, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading registry: %v", err)
	}
	if _, err := Assign([]string{"bond"}); err != nil {
		t.Fatalf("Assign (repeat): %v", err)
	}
	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("re-reading registry: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("a no-op Assign rewrote the registry:\nbefore:\n%s\nafter:\n%s",
			before, after)
	}
}

func TestEmojiForRespectsAHandEditedMark(t *testing.T) {
	// The registry is a plain TOML file and an operator is invited to choose their
	// own marks. An assignment pass must read what they wrote, not overwrite it.
	isolate(t)
	p, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if err := os.WriteFile(p, []byte("[agents]\n\"q\" = \"🎩\"\n"), 0o600); err != nil {
		t.Fatalf("seeding registry: %v", err)
	}
	m, err := Assign([]string{"q", "bond"})
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if m["q"] != "🎩" {
		t.Errorf("hand-edited mark for q was replaced with %q", m["q"])
	}
	if m["bond"] == "🎩" {
		t.Error("bond was given the mark q already holds")
	}
	if got := EmojiFor("q"); got != "🎩" {
		t.Errorf("EmojiFor(q) = %q, want the hand-edited 🎩", got)
	}
}

func TestEmojiForIsEmptyWhenUnassigned(t *testing.T) {
	// "" is a real answer: the palette is finite and a duplicate mark would be
	// worse than none, so callers render the name alone.
	isolate(t)
	if got := EmojiFor("nobody"); got != "" {
		t.Errorf("EmojiFor(nobody) = %q, want empty", got)
	}
}

func TestPaletteEntriesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range Palette {
		if seen[e] {
			t.Errorf("palette contains %q twice — two agents could be badged alike", e)
		}
		seen[e] = true
	}
	// The bar already spends these glyphs on meaning; a mark that reads as a
	// status would be worse than no mark.
	for _, reserved := range []string{"⚓", "⚙", "⚠", "✉", "⏱", "Σ", "⚑"} {
		if seen[reserved] {
			t.Errorf("palette reuses %q, which the bar uses to mean something", reserved)
		}
	}
}
