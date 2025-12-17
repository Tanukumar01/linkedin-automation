package storage

import (
	"time"
)

// ConnectionRequest represents a sent connection request
type ConnectionRequest struct {
	ID          int64
	ProfileURL  string
	ProfileName string
	JobTitle    string
	Company     string
	Note        string
	Status      string // pending, accepted, rejected, withdrawn
	SentAt      time.Time
	UpdatedAt   time.Time
}

// Message represents a sent message
type Message struct {
	ID          int64
	ProfileURL  string
	ProfileName string
	Content     string
	SentAt      time.Time
}

// SearchResult represents a cached search result
type SearchResult struct {
	ID          int64
	ProfileURL  string
	ProfileName string
	JobTitle    string
	Company     string
	Location    string
	FoundAt     time.Time
	Contacted   bool
}

// ActivityLog represents a logged activity
type ActivityLog struct {
	ID        int64
	Action    string // login, search, connect, message, etc.
	Details   string
	Timestamp time.Time
}

// DailyStats represents daily activity statistics
type DailyStats struct {
	Date              string
	ConnectionsSent   int
	ConnectionsAccepted int
	MessagesSent      int
	SearchesPerformed int
}
