package session

import "testing"

func TestStateWord(t *testing.T) {
	cases := map[string]string{
		"busy":           "busy",
		"idle":           "idle",
		"waiting":        "waiting",
		"saturated·948k": "saturated", // a stat rides in the cell; the verdict is still saturated
		"busy+1sh":       "busy",
		"?":              "?",
		"":               "",
		"worker":         "", // a role is not a verdict
		"current":        "", // a currency is not a verdict
	}
	for cell, want := range cases {
		if got := stateWord(cell); got != want {
			t.Errorf("stateWord(%q) = %q, want %q", cell, got, want)
		}
	}
}

func TestParseCrew(t *testing.T) {
	// Real `st crew` shape: name role up currency state pane. Position-tolerant
	// parse must pick name, the verdict cell, and the currency regardless.
	out := `
  arnold      worker         up       unknown  saturated·948k   sess-arnold
  billy       worker         up       STALE    busy             sess-billy
  franklin    worker         up       STALE    busy+1sh         sess-franklin
  kelly       worker         up       current  idle             sess-kelly
  noise line with no verdict at all
`
	crew := parseCrew(out)
	if len(crew) != 4 {
		t.Fatalf("parseCrew got %d entries, want 4 (noise line dropped): %+v", len(crew), crew)
	}
	if crew["arnold"].State != "saturated·948k" || crew["arnold"].Currency != "unknown" {
		t.Errorf("arnold = %+v", crew["arnold"])
	}
	if crew["billy"].State != "busy" || crew["billy"].Currency != "STALE" {
		t.Errorf("billy = %+v", crew["billy"])
	}
	if crew["kelly"].Currency != "current" {
		t.Errorf("kelly currency = %q, want current", crew["kelly"].Currency)
	}
}

func TestBuildRowsSortsAttentionFirst(t *testing.T) {
	sessions := []string{"kelly", "arnold", "dearing", "billy", "main"}
	crew := map[string]crewEntry{
		"kelly":   {State: "idle", Currency: "current"},
		"arnold":  {State: "saturated·948k", Currency: "unknown"},
		"dearing": {State: "waiting", Currency: "unknown"},
		"billy":   {State: "busy", Currency: "STALE"},
		// "main" is a live session st does not know as crew.
	}
	items := map[string]string{"dearing": "item-9", "arnold": "item-2"}
	rows := buildRows(sessions, crew, func(n string) string { return items[n] })

	// Expected order: waiting, saturated, busy, idle, then the unknown session.
	wantOrder := []string{"dearing", "arnold", "billy", "kelly", "main"}
	if len(rows) != len(wantOrder) {
		t.Fatalf("got %d rows, want %d", len(rows), len(wantOrder))
	}
	for i, w := range wantOrder {
		if rows[i].Name != w {
			names := make([]string, len(rows))
			for j, r := range rows {
				names[j] = r.Name
			}
			t.Fatalf("order = %v, want %v", names, wantOrder)
		}
	}
	// The item lookup is joined onto the right row.
	if rows[0].Name != "dearing" || rows[0].Item != "item-9" {
		t.Errorf("dearing row = %+v, want Item item-9", rows[0])
	}
	// A session st does not know as crew still lists, with no invented verdict.
	last := rows[len(rows)-1]
	if last.Name != "main" || last.State != "" {
		t.Errorf("unknown session row = %+v, want empty State", last)
	}
}

// The ROLE column exists because an operator could not tell which pane was the
// coordinator's — the question that prompted it was literally "are you
// administrator?". The role must ride the
// SAME row as its agent — a role joined onto the wrong row is worse than no
// column, because it tells the operator confidently who to escalate to and is
// wrong. Sorting reorders rows, so this asserts role-to-name AFTER the sort.
func TestBuildRowsCarriesRolePerAgent(t *testing.T) {
	sessions := []string{"kelly", "sattler", "billy", "main"}
	crew := map[string]crewEntry{
		"kelly":   {Role: "worker", State: "idle", Currency: "current"},
		"sattler": {Role: "administrator", State: "waiting", Currency: "current"},
		"billy":   {Role: "lead", State: "busy", Currency: "STALE"},
		// "main" is a live session st does not know as crew — no role to invent.
	}
	rows := buildRows(sessions, crew, func(string) string { return "" })

	want := map[string]string{
		"kelly": "worker", "sattler": "administrator", "billy": "lead", "main": "",
	}
	for _, r := range rows {
		if got := r.Role; got != want[r.Name] {
			t.Errorf("%s Role = %q, want %q", r.Name, got, want[r.Name])
		}
	}
	// The role is carried in FULL. The bar abbreviates (shortRole) because it has
	// a 30-column budget; the table does not, and truncating here would drop the
	// distinction the column was added to show.
	for _, r := range rows {
		if r.Name == "sattler" && r.Role != "administrator" {
			t.Errorf("sattler Role = %q, want the unabbreviated \"administrator\"", r.Role)
		}
	}
}

func TestBuildRowsStableWithinRank(t *testing.T) {
	// Two agents at the same rank sort by name, deterministically.
	sessions := []string{"zeb", "abe"}
	crew := map[string]crewEntry{
		"zeb": {State: "busy"},
		"abe": {State: "busy"},
	}
	rows := buildRows(sessions, crew, func(string) string { return "" })
	if rows[0].Name != "abe" || rows[1].Name != "zeb" {
		t.Errorf("same-rank order = %q,%q; want abe,zeb", rows[0].Name, rows[1].Name)
	}
}
