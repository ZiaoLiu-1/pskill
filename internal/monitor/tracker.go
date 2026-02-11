package monitor

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Event struct {
	SkillName string
	CLI       string
	Project   string
	EventType string
	Timestamp time.Time
}

type Aggregates struct {
	TopSkills map[string]int64
	ByCLI     map[string]int64
	Recent    []Event
	Stale     []string
}

type Tracker struct {
	db *sql.DB
}

func NewTracker(dbPath string) (*Tracker, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	t := &Tracker{db: db}
	return t, t.migrate()
}

func (t *Tracker) migrate() error {
	_, err := t.db.Exec(`
CREATE TABLE IF NOT EXISTS usage (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  skill_name TEXT NOT NULL,
  cli TEXT NOT NULL,
  project TEXT NOT NULL,
  event_type TEXT NOT NULL,
  timestamp DATETIME NOT NULL
);`)
	return err
}

func (t *Tracker) Record(ev Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}
	_, err := t.db.Exec(
		`INSERT INTO usage(skill_name, cli, project, event_type, timestamp) VALUES (?, ?, ?, ?, ?)`,
		ev.SkillName, ev.CLI, ev.Project, ev.EventType, ev.Timestamp.UTC(),
	)
	return err
}

func (t *Tracker) Stats() (Aggregates, error) {
	out := Aggregates{
		TopSkills: map[string]int64{},
		ByCLI:     map[string]int64{},
		Recent:    []Event{},
		Stale:     []string{},
	}

	rows1, err := t.db.Query(`SELECT skill_name, COUNT(*) c FROM usage GROUP BY skill_name ORDER BY c DESC LIMIT 10`)
	if err == nil {
		for rows1.Next() {
			var n string
			var c int64
			_ = rows1.Scan(&n, &c)
			out.TopSkills[n] = c
		}
		rows1.Close()
	}

	rows2, err := t.db.Query(`SELECT cli, COUNT(*) c FROM usage GROUP BY cli ORDER BY c DESC`)
	if err == nil {
		for rows2.Next() {
			var n string
			var c int64
			_ = rows2.Scan(&n, &c)
			out.ByCLI[n] = c
		}
		rows2.Close()
	}

	rows3, err := t.db.Query(`SELECT skill_name, cli, project, event_type, timestamp FROM usage ORDER BY timestamp DESC LIMIT 20`)
	if err == nil {
		for rows3.Next() {
			ev := Event{}
			_ = rows3.Scan(&ev.SkillName, &ev.CLI, &ev.Project, &ev.EventType, &ev.Timestamp)
			out.Recent = append(out.Recent, ev)
		}
		rows3.Close()
	}

	rows4, err := t.db.Query(`
SELECT skill_name
FROM usage
GROUP BY skill_name
HAVING MAX(timestamp) < DATETIME('now', '-60 day')
ORDER BY MAX(timestamp) ASC`)
	if err == nil {
		for rows4.Next() {
			var name string
			_ = rows4.Scan(&name)
			out.Stale = append(out.Stale, name)
		}
		rows4.Close()
	}

	return out, nil
}

func (t *Tracker) Close() error {
	return t.db.Close()
}
