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
	if crew["arnold"].state != "saturated·948k" || crew["arnold"].currency != "unknown" {
		t.Errorf("arnold = %+v", crew["arnold"])
	}
	if crew["billy"].state != "busy" || crew["billy"].currency != "STALE" {
		t.Errorf("billy = %+v", crew["billy"])
	}
	if crew["kelly"].currency != "current" {
		t.Errorf("kelly currency = %q, want current", crew["kelly"].currency)
	}
}

func TestBuildRowsSortsAttentionFirst(t *testing.T) {
	sessions := []string{"kelly", "arnold", "dearing", "billy", "main"}
	crew := map[string]crewEntry{
		"kelly":   {state: "idle", currency: "current"},
		"arnold":  {state: "saturated·948k", currency: "unknown"},
		"dearing": {state: "waiting", currency: "unknown"},
		"billy":   {state: "busy", currency: "STALE"},
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

func TestBuildRowsStableWithinRank(t *testing.T) {
	// Two agents at the same rank sort by name, deterministically.
	sessions := []string{"zeb", "abe"}
	crew := map[string]crewEntry{
		"zeb": {state: "busy"},
		"abe": {state: "busy"},
	}
	rows := buildRows(sessions, crew, func(string) string { return "" })
	if rows[0].Name != "abe" || rows[1].Name != "zeb" {
		t.Errorf("same-rank order = %q,%q; want abe,zeb", rows[0].Name, rows[1].Name)
	}
}
