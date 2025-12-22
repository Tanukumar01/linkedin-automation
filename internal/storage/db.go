package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Run migrations
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate runs database migrations
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS connection_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_url TEXT NOT NULL UNIQUE,
			profile_name TEXT,
			job_title TEXT,
			company TEXT,
			note TEXT,
			status TEXT DEFAULT 'pending',
			sent_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_url TEXT NOT NULL,
			profile_name TEXT,
			content TEXT NOT NULL,
			sent_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS search_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_url TEXT NOT NULL UNIQUE,
			profile_name TEXT,
			job_title TEXT,
			company TEXT,
			location TEXT,
			found_at DATETIME NOT NULL,
			contacted BOOLEAN DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			details TEXT,
			timestamp DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_requests_status ON connection_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_requests_sent_at ON connection_requests(sent_at)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at)`,
		`CREATE INDEX IF NOT EXISTS idx_search_results_contacted ON search_results(contacted)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// SaveConnectionRequest saves a connection request to the database
func (db *DB) SaveConnectionRequest(req *ConnectionRequest) error {
	query := `INSERT INTO connection_requests (profile_url, profile_name, job_title, company, note, status, sent_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.conn.Exec(query, req.ProfileURL, req.ProfileName, req.JobTitle, req.Company, req.Note, req.Status, req.SentAt, req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save connection request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	req.ID = id
	return nil
}

// UpdateConnectionStatus updates the status of a connection request
func (db *DB) UpdateConnectionStatus(profileURL, status string) error {
	query := `UPDATE connection_requests SET status = ?, updated_at = ? WHERE profile_url = ?`
	_, err := db.conn.Exec(query, status, time.Now(), profileURL)
	return err
}

// GetConnectionRequestsByDate returns connection requests sent on a specific date
func (db *DB) GetConnectionRequestsByDate(date time.Time) ([]ConnectionRequest, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT id, profile_url, profile_name, job_title, company, note, status, sent_at, updated_at
			  FROM connection_requests WHERE sent_at >= ? AND sent_at < ?`

	rows, err := db.conn.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []ConnectionRequest
	for rows.Next() {
		var req ConnectionRequest
		if err := rows.Scan(&req.ID, &req.ProfileURL, &req.ProfileName, &req.JobTitle, &req.Company, &req.Note, &req.Status, &req.SentAt, &req.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetConnectionRequestsCountByDate returns the count of connection requests sent on a specific date
func (db *DB) GetConnectionRequestsCountByDate(date time.Time) (int, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT COUNT(*) FROM connection_requests WHERE sent_at >= ? AND sent_at < ?`

	var count int
	err := db.conn.QueryRow(query, startOfDay, endOfDay).Scan(&count)
	return count, err
}

// IsProfileContacted checks if a profile has already been contacted
func (db *DB) IsProfileContacted(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?`

	var count int
	err := db.conn.QueryRow(query, profileURL).Scan(&count)
	return count > 0, err
}

// SaveMessage saves a message to the database
func (db *DB) SaveMessage(msg *Message) error {
	query := `INSERT INTO messages (profile_url, profile_name, content, sent_at)
			  VALUES (?, ?, ?, ?)`

	result, err := db.conn.Exec(query, msg.ProfileURL, msg.ProfileName, msg.Content, msg.SentAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	msg.ID = id
	return nil
}

// GetMessagesCountByDate returns the count of messages sent on a specific date
func (db *DB) GetMessagesCountByDate(date time.Time) (int, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT COUNT(*) FROM messages WHERE sent_at >= ? AND sent_at < ?`

	var count int
	err := db.conn.QueryRow(query, startOfDay, endOfDay).Scan(&count)
	return count, err
}

// SaveSearchResult saves a search result to the database
func (db *DB) SaveSearchResult(result *SearchResult) error {
	query := `INSERT OR IGNORE INTO search_results (profile_url, profile_name, job_title, company, location, found_at, contacted)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := db.conn.Exec(query, result.ProfileURL, result.ProfileName, result.JobTitle, result.Company, result.Location, result.FoundAt, result.Contacted)
	if err != nil {
		return fmt.Errorf("failed to save search result: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		result.ID = id
	}

	return nil
}

// GetUncontactedProfiles returns profiles that haven't been contacted yet
func (db *DB) GetUncontactedProfiles(limit int) ([]SearchResult, error) {
	query := `SELECT id, profile_url, profile_name, job_title, company, location, found_at, contacted
			  FROM search_results WHERE contacted = 0 LIMIT ?`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.ProfileURL, &result.ProfileName, &result.JobTitle, &result.Company, &result.Location, &result.FoundAt, &result.Contacted); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// MarkProfileContacted marks a profile as contacted
func (db *DB) MarkProfileContacted(profileURL string) error {
	query := `UPDATE search_results SET contacted = 1 WHERE profile_url = ?`
	_, err := db.conn.Exec(query, profileURL)
	return err
}

// LogActivity logs an activity to the database
func (db *DB) LogActivity(action, details string) error {
	query := `INSERT INTO activity_logs (action, details, timestamp) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, action, details, time.Now())
	return err
}

// GetDailyStats returns statistics for a specific date
func (db *DB) GetDailyStats(date time.Time) (*DailyStats, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats := &DailyStats{
		Date: date.Format("2006-01-02"),
	}

	// Count connections sent
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM connection_requests WHERE sent_at >= ? AND sent_at < ?`, startOfDay, endOfDay).Scan(&stats.ConnectionsSent)
	if err != nil {
		return nil, err
	}

	// Count connections accepted
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM connection_requests WHERE status = 'accepted' AND updated_at >= ? AND updated_at < ?`, startOfDay, endOfDay).Scan(&stats.ConnectionsAccepted)
	if err != nil {
		return nil, err
	}

	// Count messages sent
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM messages WHERE sent_at >= ? AND sent_at < ?`, startOfDay, endOfDay).Scan(&stats.MessagesSent)
	if err != nil {
		return nil, err
	}

	// Count searches performed
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM activity_logs WHERE action = 'search' AND timestamp >= ? AND timestamp < ?`, startOfDay, endOfDay).Scan(&stats.SearchesPerformed)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
