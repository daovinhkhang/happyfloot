package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func (s *runStore) hasAnyData() (bool, error) {
	var total int64
	err := s.queryRow(
		`SELECT
			(SELECT COUNT(*) FROM users) +
			(SELECT COUNT(*) FROM runs) +
			(SELECT COUNT(*) FROM run_logs) +
			(SELECT COUNT(*) FROM sessions)`,
	).Scan(&total)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func sqliteTableExists(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?`, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func resetSequence(tx *sql.Tx, table string) error {
	query := fmt.Sprintf(
		`SELECT setval(pg_get_serial_sequence('%s', 'id'),
			COALESCE((SELECT MAX(id) FROM %s), 1),
			(SELECT COUNT(*) > 0 FROM %s))`,
		table, table, table,
	)
	_, err := tx.Exec(query)
	return err
}

func removeSQLiteFiles(path string) {
	for _, candidate := range []string{path, path + "-shm", path + "-wal"} {
		if candidate == "" {
			continue
		}
		if err := os.Remove(candidate); err == nil {
			fmt.Println("Removed legacy SQLite file:", candidate)
		}
	}
}

func (s *runStore) ImportFromSQLite(path string, deleteAfter bool) error {
	sqlitePath := strings.TrimSpace(path)
	if sqlitePath == "" {
		return nil
	}
	if _, err := os.Stat(sqlitePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("SQLite migration source not found, skipping:", sqlitePath)
			return nil
		}
		return err
	}
	nonEmpty, err := s.hasAnyData()
	if err != nil {
		return err
	}
	if nonEmpty {
		fmt.Println("PostgreSQL already has data; skipping SQLite migration.")
		if deleteAfter {
			removeSQLiteFiles(sqlitePath)
		}
		return nil
	}

	oldDB, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		return err
	}
	defer oldDB.Close()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	importedUsers := 0
	importedRuns := 0
	importedLogs := 0
	importedSessions := 0

	usersTable, err := sqliteTableExists(oldDB, "users")
	if err != nil {
		return err
	}
	if usersTable {
		rows, err := oldDB.Query(
			`SELECT id, username, password_hash, role, can_start_run, can_view_monitor, is_active, created_at, updated_at, created_by
			 FROM users
			 ORDER BY id ASC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var username string
			var passwordHash string
			var role string
			var canStart int
			var canView int
			var isActive int
			var createdAt string
			var updatedAt string
			var createdBy sql.NullInt64
			if err := rows.Scan(&id, &username, &passwordHash, &role, &canStart, &canView, &isActive, &createdAt, &updatedAt, &createdBy); err != nil {
				return err
			}
			var createdByValue interface{}
			if createdBy.Valid {
				createdByValue = createdBy.Int64
			}
			if _, err := tx.Exec(
				rebindQuestionToDollar(
					`INSERT INTO users (id, username, password_hash, role, can_start_run, can_view_monitor, is_active, created_at, updated_at, created_by)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
					 ON CONFLICT (id) DO NOTHING`,
				),
				id, username, passwordHash, role, canStart, canView, isActive, createdAt, updatedAt, createdByValue,
			); err != nil {
				return err
			}
			importedUsers++
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	runsTable, err := sqliteTableExists(oldDB, "runs")
	if err != nil {
		return err
	}
	if runsTable {
		rows, err := oldDB.Query(
			`SELECT id, target_url, threads, requests_per_conn, method, seconds, header_preset, header_text, status, error,
			        created_at, started_at, completed_at, updated_at
			 FROM runs
			 ORDER BY id ASC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var targetURL string
			var threads int
			var requestsPerConn int
			var method string
			var seconds int
			var headerPreset string
			var headerText string
			var status string
			var runErr string
			var createdAt string
			var startedAt sql.NullString
			var completedAt sql.NullString
			var updatedAt string
			if err := rows.Scan(
				&id, &targetURL, &threads, &requestsPerConn, &method, &seconds, &headerPreset, &headerText,
				&status, &runErr, &createdAt, &startedAt, &completedAt, &updatedAt,
			); err != nil {
				return err
			}
			var startedAtValue interface{}
			if startedAt.Valid {
				startedAtValue = startedAt.String
			}
			var completedAtValue interface{}
			if completedAt.Valid {
				completedAtValue = completedAt.String
			}
			if _, err := tx.Exec(
				rebindQuestionToDollar(
					`INSERT INTO runs (id, target_url, threads, requests_per_conn, method, seconds, header_preset, header_text, status, error, created_at, started_at, completed_at, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
					 ON CONFLICT (id) DO NOTHING`,
				),
				id, targetURL, threads, requestsPerConn, method, seconds, headerPreset, headerText, status, runErr,
				createdAt, startedAtValue, completedAtValue, updatedAt,
			); err != nil {
				return err
			}
			importedRuns++
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	logsTable, err := sqliteTableExists(oldDB, "run_logs")
	if err != nil {
		return err
	}
	if logsTable {
		rows, err := oldDB.Query(
			`SELECT id, run_id, layer, message, created_at
			 FROM run_logs
			 ORDER BY id ASC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var runID int64
			var layer string
			var message string
			var createdAt string
			if err := rows.Scan(&id, &runID, &layer, &message, &createdAt); err != nil {
				return err
			}
			if _, err := tx.Exec(
				rebindQuestionToDollar(
					`INSERT INTO run_logs (id, run_id, layer, message, created_at)
					 VALUES (?, ?, ?, ?, ?)
					 ON CONFLICT (id) DO NOTHING`,
				),
				id, runID, layer, message, createdAt,
			); err != nil {
				return err
			}
			importedLogs++
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	sessionsTable, err := sqliteTableExists(oldDB, "sessions")
	if err != nil {
		return err
	}
	if sessionsTable {
		rows, err := oldDB.Query(
			`SELECT id, user_id, expires_at, created_at, last_seen_at
			 FROM sessions`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			var userID int64
			var expiresAt string
			var createdAt string
			var lastSeenAt string
			if err := rows.Scan(&id, &userID, &expiresAt, &createdAt, &lastSeenAt); err != nil {
				return err
			}
			if _, err := tx.Exec(
				rebindQuestionToDollar(
					`INSERT INTO sessions (id, user_id, expires_at, created_at, last_seen_at)
					 VALUES (?, ?, ?, ?, ?)
					 ON CONFLICT (id) DO NOTHING`,
				),
				id, userID, expiresAt, createdAt, lastSeenAt,
			); err != nil {
				return err
			}
			importedSessions++
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	for _, table := range []string{"users", "runs", "run_logs"} {
		if err := resetSequence(tx, table); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	fmt.Printf(
		"SQLite migration completed from %s (users=%d, runs=%d, run_logs=%d, sessions=%d)\n",
		sqlitePath, importedUsers, importedRuns, importedLogs, importedSessions,
	)
	if deleteAfter {
		removeSQLiteFiles(sqlitePath)
	}
	return nil
}
