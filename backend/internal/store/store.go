// Package store persists sessions, their versioned filter/rubric/result
// history, chat messages, and liked signals to SQLite.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	_ "modernc.org/sqlite"

	"talent-search-rubric-arm/backend/internal/domain"
)

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	original_query TEXT NOT NULL,
	created_at TEXT NOT NULL,
	status TEXT NOT NULL,
	frozen_version_number INTEGER
);

CREATE TABLE IF NOT EXISTS versions (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	version_number INTEGER NOT NULL,
	trigger_type TEXT NOT NULL,
	trigger_text TEXT,
	filters_json TEXT NOT NULL,
	rubric_json TEXT NOT NULL,
	change_summary TEXT,
	results_json TEXT NOT NULL,
	empty_diagnostic_json TEXT,
	created_at TEXT NOT NULL,
	UNIQUE(session_id, version_number)
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	version_id TEXT REFERENCES versions(id),
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS liked_signals (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	version_id TEXT NOT NULL REFERENCES versions(id),
	profile_id TEXT NOT NULL,
	criterion_id TEXT NOT NULL,
	criterion_label TEXT NOT NULL,
	evidence_text TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1) // modernc.org/sqlite + concurrent writers is not worth the complexity here

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateSession(ctx context.Context, originalQuery string) (domain.Session, error) {
	sess := domain.Session{
		ID:            uuid.NewString(),
		OriginalQuery: originalQuery,
		CreatedAt:     time.Now().UTC(),
		Status:        domain.StatusActive,
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, original_query, created_at, status) VALUES (?, ?, ?, ?)`,
		sess.ID, sess.OriginalQuery, sess.CreatedAt.Format(time.RFC3339Nano), sess.Status,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("inserting session: %w", err)
	}
	return sess, nil
}

func (s *Store) GetSession(ctx context.Context, id string) (domain.Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, original_query, created_at, status, frozen_version_number FROM sessions WHERE id = ?`, id)

	var sess domain.Session
	var createdAt string
	var frozenVN sql.NullInt64
	if err := row.Scan(&sess.ID, &sess.OriginalQuery, &createdAt, &sess.Status, &frozenVN); err != nil {
		if err == sql.ErrNoRows {
			return domain.Session{}, fmt.Errorf("session %s not found: %w", id, err)
		}
		return domain.Session{}, err
	}
	sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	if frozenVN.Valid {
		v := int(frozenVN.Int64)
		sess.FrozenVersionNumber = &v
	}
	return sess, nil
}

func (s *Store) ListSessions(ctx context.Context) ([]domain.SessionSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, original_query, created_at, status FROM sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.SessionSummary{}
	for rows.Next() {
		var sum domain.SessionSummary
		var createdAt string
		if err := rows.Scan(&sum.ID, &sum.OriginalQuery, &createdAt, &sum.Status); err != nil {
			return nil, err
		}
		sum.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, sum)
	}
	return out, rows.Err()
}

func (s *Store) FreezeSession(ctx context.Context, sessionID string, versionNumber int) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET status = ?, frozen_version_number = ? WHERE id = ?`,
		domain.StatusFrozen, versionNumber, sessionID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s not found", sessionID)
	}
	return nil
}

func (s *Store) NextVersionNumber(ctx context.Context, sessionID string) (int, error) {
	var max sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT MAX(version_number) FROM versions WHERE session_id = ?`, sessionID).Scan(&max)
	if err != nil {
		return 0, err
	}
	if !max.Valid {
		return 1, nil
	}
	return int(max.Int64) + 1, nil
}

func (s *Store) CreateVersion(ctx context.Context, v domain.Version) (domain.Version, error) {
	v.ID = uuid.NewString()
	v.CreatedAt = time.Now().UTC()

	filtersJSON, err := json.Marshal(v.Filters)
	if err != nil {
		return domain.Version{}, err
	}
	rubricJSON, err := json.Marshal(v.Rubric)
	if err != nil {
		return domain.Version{}, err
	}
	resultsJSON, err := json.Marshal(v.Results)
	if err != nil {
		return domain.Version{}, err
	}
	var emptyDiagJSON []byte
	if v.EmptyDiagnostic != nil {
		emptyDiagJSON, err = json.Marshal(v.EmptyDiagnostic)
		if err != nil {
			return domain.Version{}, err
		}
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO versions (id, session_id, version_number, trigger_type, trigger_text, filters_json, rubric_json, change_summary, results_json, empty_diagnostic_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.SessionID, v.VersionNumber, v.TriggerType, nullIfEmpty(v.TriggerText),
		string(filtersJSON), string(rubricJSON), nullIfEmpty(v.ChangeSummary), string(resultsJSON),
		nullBytesIfEmpty(emptyDiagJSON), v.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.Version{}, fmt.Errorf("inserting version: %w", err)
	}
	return v, nil
}

func (s *Store) ListVersions(ctx context.Context, sessionID string) ([]domain.Version, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, version_number, trigger_type, trigger_text, filters_json, rubric_json, change_summary, results_json, empty_diagnostic_json, created_at
		 FROM versions WHERE session_id = ? ORDER BY version_number ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Version{}
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) LatestVersion(ctx context.Context, sessionID string) (domain.Version, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, session_id, version_number, trigger_type, trigger_text, filters_json, rubric_json, change_summary, results_json, empty_diagnostic_json, created_at
		 FROM versions WHERE session_id = ? ORDER BY version_number DESC LIMIT 1`, sessionID)
	return scanVersion(row)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVersion(row scannable) (domain.Version, error) {
	var v domain.Version
	var triggerText, changeSummary, emptyDiagJSON sql.NullString
	var filtersJSON, rubricJSON, resultsJSON, createdAt string

	if err := row.Scan(&v.ID, &v.SessionID, &v.VersionNumber, &v.TriggerType, &triggerText,
		&filtersJSON, &rubricJSON, &changeSummary, &resultsJSON, &emptyDiagJSON, &createdAt); err != nil {
		return domain.Version{}, err
	}

	v.TriggerText = triggerText.String
	v.ChangeSummary = changeSummary.String
	v.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)

	if err := json.Unmarshal([]byte(filtersJSON), &v.Filters); err != nil {
		return domain.Version{}, err
	}
	if err := json.Unmarshal([]byte(rubricJSON), &v.Rubric); err != nil {
		return domain.Version{}, err
	}
	if err := json.Unmarshal([]byte(resultsJSON), &v.Results); err != nil {
		return domain.Version{}, err
	}
	if emptyDiagJSON.Valid && emptyDiagJSON.String != "" {
		var d domain.EmptyResultDiagnostic
		if err := json.Unmarshal([]byte(emptyDiagJSON.String), &d); err != nil {
			return domain.Version{}, err
		}
		v.EmptyDiagnostic = &d
	}

	return v, nil
}

func (s *Store) AddChatMessage(ctx context.Context, msg domain.ChatMessage) (domain.ChatMessage, error) {
	msg.ID = uuid.NewString()
	msg.CreatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chat_messages (id, session_id, version_id, role, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.SessionID, nullIfEmpty(msg.VersionID), msg.Role, msg.Content, msg.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.ChatMessage{}, err
	}
	return msg, nil
}

func (s *Store) ListChatMessages(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, version_id, role, content, created_at FROM chat_messages WHERE session_id = ? ORDER BY created_at ASC`,
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ChatMessage{}
	for rows.Next() {
		var m domain.ChatMessage
		var versionID sql.NullString
		var createdAt string
		if err := rows.Scan(&m.ID, &m.SessionID, &versionID, &m.Role, &m.Content, &createdAt); err != nil {
			return nil, err
		}
		m.VersionID = versionID.String
		m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) AddLikedSignal(ctx context.Context, sig domain.LikedSignal) (domain.LikedSignal, error) {
	sig.ID = uuid.NewString()
	sig.CreatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO liked_signals (id, session_id, version_id, profile_id, criterion_id, criterion_label, evidence_text, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sig.ID, sig.SessionID, sig.VersionID, sig.ProfileID, sig.CriterionID, sig.CriterionLabel, sig.EvidenceText,
		sig.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.LikedSignal{}, err
	}
	return sig, nil
}

func (s *Store) RemoveLikedSignal(ctx context.Context, sessionID, signalID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM liked_signals WHERE id = ? AND session_id = ?`, signalID, sessionID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("liked signal %s not found in session %s", signalID, sessionID)
	}
	return nil
}

func (s *Store) ListLikedSignals(ctx context.Context, sessionID string) ([]domain.LikedSignal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, version_id, profile_id, criterion_id, criterion_label, evidence_text, created_at
		 FROM liked_signals WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.LikedSignal{}
	for rows.Next() {
		var sig domain.LikedSignal
		var createdAt string
		if err := rows.Scan(&sig.ID, &sig.SessionID, &sig.VersionID, &sig.ProfileID, &sig.CriterionID,
			&sig.CriterionLabel, &sig.EvidenceText, &createdAt); err != nil {
			return nil, err
		}
		sig.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, sig)
	}
	return out, rows.Err()
}

func (s *Store) GetSessionDetail(ctx context.Context, sessionID string) (domain.SessionDetail, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	versions, err := s.ListVersions(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	chats, err := s.ListChatMessages(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	signals, err := s.ListLikedSignals(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	return domain.SessionDetail{Session: sess, Versions: versions, ChatMessages: chats, LikedSignals: signals}, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullBytesIfEmpty(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
