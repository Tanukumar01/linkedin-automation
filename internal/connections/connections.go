package connections

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"github.com/Tanukumar01/linkedin-automation/internal/config"
	"github.com/Tanukumar01/linkedin-automation/internal/logger"
	"github.com/Tanukumar01/linkedin-automation/internal/stealth"
	"github.com/Tanukumar01/linkedin-automation/internal/storage"
)

// ConnectionManager handles connection requests
type ConnectionManager struct {
	page     *rod.Page
	config   *config.ConnectionsConfig
	db       *storage.DB
	timing   *stealth.TimingController
	typer    *stealth.Typer
	mouse    *stealth.MouseMover
	scroller *stealth.Scroller
	rand     *rand.Rand
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(page *rod.Page, cfg *config.ConnectionsConfig, db *storage.DB, timing *stealth.TimingController, typer *stealth.Typer, mouse *stealth.MouseMover, scroller *stealth.Scroller) *ConnectionManager {
	return &ConnectionManager{
		page:     page,
		config:   cfg,
		db:       db,
		timing:   timing,
		typer:    typer,
		mouse:    mouse,
		scroller: scroller,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SendConnectionRequest sends a connection request to a profile
func (cm *ConnectionManager) SendConnectionRequest(profileURL, profileName, jobTitle, company string) error {
	logger.Infof("Sending connection request to: %s", profileName)

	// Check daily limit
	if err := cm.checkDailyLimit(); err != nil {
		return err
	}

	// Check if already contacted
	contacted, err := cm.db.IsProfileContacted(profileURL)
	if err != nil {
		return fmt.Errorf("failed to check if profile contacted: %w", err)
	}

	if contacted {
		logger.Infof("Profile already contacted: %s", profileName)
		return nil
	}

	// Navigate to profile
	if err := cm.page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	if err := cm.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for profile page: %w", err)
	}

	cm.timing.Wait(cm.timing.ThinkTime())

	// Scroll to view profile
	if err := cm.scroller.ScrollDown(cm.page, 300); err != nil {
		logger.Warnf("Failed to scroll: %v", err)
	}

	cm.timing.Wait(cm.timing.ShortPause())

	// Find Connect button
	connectButton, err := cm.findConnectButton()
	if err != nil {
		return fmt.Errorf("failed to find connect button: %w", err)
	}

	// Click Connect button with human-like mouse movement
	if err := cm.mouse.ClickElement(connectButton); err != nil {
		return fmt.Errorf("failed to click connect button: %w", err)
	}

	cm.timing.Wait(cm.timing.ShortPause())

	// Check if "Add a note" option is available
	hasNoteOption := cm.hasAddNoteOption()

	var note string
	if hasNoteOption {
		// Click "Add a note" button
		if err := cm.clickAddNoteButton(); err != nil {
			logger.Warnf("Failed to click add note button: %v", err)
		} else {
			cm.timing.Wait(cm.timing.ShortPause())

			// Generate personalized note
			note = cm.generateNote(profileName, jobTitle, company)

			// Type note
			if err := cm.typeNote(note); err != nil {
				logger.Warnf("Failed to type note: %v", err)
			}

			cm.timing.Wait(cm.timing.ThinkTime())
		}
	}

	// Click Send button
	if err := cm.clickSendButton(); err != nil {
		return fmt.Errorf("failed to click send button: %w", err)
	}

	logger.Infof("Connection request sent to: %s", profileName)

	// Save to database
	request := &storage.ConnectionRequest{
		ProfileURL:  profileURL,
		ProfileName: profileName,
		JobTitle:    jobTitle,
		Company:     company,
		Note:        note,
		Status:      "pending",
		SentAt:      time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := cm.db.SaveConnectionRequest(request); err != nil {
		logger.Errorf("Failed to save connection request: %v", err)
	}

	// Mark profile as contacted
	if err := cm.db.MarkProfileContacted(profileURL); err != nil {
		logger.Errorf("Failed to mark profile as contacted: %v", err)
	}

	// Log activity
	cm.db.LogActivity("connection_request", fmt.Sprintf("Sent to %s", profileName))

	// Cooldown
	cooldown := time.Duration(cm.config.CooldownBetweenRequestsMin+cm.rand.Intn(cm.config.CooldownBetweenRequestsMax-cm.config.CooldownBetweenRequestsMin+1)) * time.Second
	cm.timing.Wait(cooldown)

	return nil
}

// checkDailyLimit checks if daily connection limit has been reached
func (cm *ConnectionManager) checkDailyLimit() error {
	count, err := cm.db.GetConnectionRequestsCountByDate(time.Now())
	if err != nil {
		return fmt.Errorf("failed to get connection count: %w", err)
	}

	if count >= cm.config.DailyLimit {
		return fmt.Errorf("daily connection limit reached (%d/%d)", count, cm.config.DailyLimit)
	}

	logger.Infof("Daily connections: %d/%d", count, cm.config.DailyLimit)
	return nil
}

// findConnectButton finds the Connect button on the profile
func (cm *ConnectionManager) findConnectButton() (*rod.Element, error) {
	// Try different methods for Connect button

	// 1. Text-based search (most reliable)
	if el, err := cm.page.ElementR("button", "(?i)^Connect$"); err == nil {
		return el, nil
	}

	// 2. Aria-label based search (often contains extra text like "Connect to Name")
	if el, err := cm.page.Element("button[aria-label*='Connect']"); err == nil {
		return el, nil
	}

	// 3. Specific profile action area
	if el, err := cm.page.Element(".pvs-profile-actions button"); err == nil {
		if text, _ := el.Text(); strings.Contains(strings.ToLower(text), "connect") {
			return el, nil
		}
	}

	return nil, fmt.Errorf("connect button not found")
}

// hasAddNoteOption checks if "Add a note" option is available
func (cm *ConnectionManager) hasAddNoteOption() bool {
	has, _, _ := cm.page.Has("button[aria-label*='Add a note']")
	return has
}

// clickAddNoteButton clicks the "Add a note" button
func (cm *ConnectionManager) clickAddNoteButton() error {
	button, err := cm.page.Element("button[aria-label*='Add a note']")
	if err != nil {
		return err
	}

	return cm.mouse.ClickElement(button)
}

// typeNote types the connection note
func (cm *ConnectionManager) typeNote(note string) error {
	// Find note textarea
	textarea, err := cm.page.Element("textarea[name='message']")
	if err != nil {
		return err
	}

	return cm.typer.TypeText(cm.page, textarea, note)
}

// clickSendButton clicks the Send button
func (cm *ConnectionManager) clickSendButton() error {
	// Try multiple ways to find the send button

	// 1. Text-based (most robust)
	if el, err := cm.page.ElementR("button", "(?i)Send"); err == nil {
		return cm.mouse.ClickElement(el)
	}

	// 2. Aria-label based
	button, err := cm.page.Element("button[aria-label*='Send']")
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	return cm.mouse.ClickElement(button)
}

// generateNote generates a personalized connection note
func (cm *ConnectionManager) generateNote(profileName, jobTitle, company string) string {
	if len(cm.config.NoteTemplates) == 0 {
		return ""
	}

	// Select random template
	template := cm.config.NoteTemplates[cm.rand.Intn(len(cm.config.NoteTemplates))]

	// Extract first name
	firstName := strings.Split(profileName, " ")[0]

	// Replace variables
	note := strings.ReplaceAll(template, "{{firstName}}", firstName)
	note = strings.ReplaceAll(note, "{{jobTitle}}", jobTitle)
	note = strings.ReplaceAll(note, "{{company}}", company)

	// Ensure note doesn't exceed character limit
	if len(note) > cm.config.NoteCharacterLimit {
		note = note[:cm.config.NoteCharacterLimit-3] + "..."
	}

	return note
}

// GetPendingConnections returns pending connection requests
func (cm *ConnectionManager) GetPendingConnections() ([]storage.ConnectionRequest, error) {
	// This would query the database for pending connections
	// For now, return empty slice
	return []storage.ConnectionRequest{}, nil
}
