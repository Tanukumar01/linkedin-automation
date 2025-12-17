package messaging

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

// MessageManager handles messaging operations
type MessageManager struct {
	page     *rod.Page
	config   *config.MessagingConfig
	db       *storage.DB
	timing   *stealth.TimingController
	typer    *stealth.Typer
	mouse    *stealth.MouseMover
	scroller *stealth.Scroller
	rand     *rand.Rand
}

// NewMessageManager creates a new message manager
func NewMessageManager(page *rod.Page, cfg *config.MessagingConfig, db *storage.DB, timing *stealth.TimingController, typer *stealth.Typer, mouse *stealth.MouseMover, scroller *stealth.Scroller) *MessageManager {
	return &MessageManager{
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

// SendMessage sends a message to a connection
func (mm *MessageManager) SendMessage(profileURL, profileName, jobTitle, company string) error {
	logger.Infof("Sending message to: %s", profileName)

	// Check daily limit
	if err := mm.checkDailyLimit(); err != nil {
		return err
	}

	// Navigate to profile
	if err := mm.page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	if err := mm.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for profile page: %w", err)
	}

	mm.timing.Wait(mm.timing.ThinkTime())

	// Find Message button
	messageButton, err := mm.findMessageButton()
	if err != nil {
		return fmt.Errorf("failed to find message button: %w", err)
	}

	// Click Message button
	if err := mm.mouse.ClickElement(messageButton); err != nil {
		return fmt.Errorf("failed to click message button: %w", err)
	}

	mm.timing.Wait(mm.timing.ShortPause())

	// Generate message
	message := mm.generateMessage(profileName, jobTitle, company)

	// Type message
	if err := mm.typeMessage(message); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	mm.timing.Wait(mm.timing.ThinkTime())

	// Send message
	if err := mm.clickSendButton(); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logger.Infof("Message sent to: %s", profileName)

	// Save to database
	msg := &storage.Message{
		ProfileURL:  profileURL,
		ProfileName: profileName,
		Content:     message,
		SentAt:      time.Now(),
	}

	if err := mm.db.SaveMessage(msg); err != nil {
		logger.Errorf("Failed to save message: %v", err)
	}

	// Log activity
	mm.db.LogActivity("message_sent", fmt.Sprintf("Sent to %s", profileName))

	// Cooldown
	cooldown := time.Duration(mm.config.CooldownBetweenMessagesMin+mm.rand.Intn(mm.config.CooldownBetweenMessagesMax-mm.config.CooldownBetweenMessagesMin+1)) * time.Second
	mm.timing.Wait(cooldown)

	return nil
}

// checkDailyLimit checks if daily message limit has been reached
func (mm *MessageManager) checkDailyLimit() error {
	count, err := mm.db.GetMessagesCountByDate(time.Now())
	if err != nil {
		return fmt.Errorf("failed to get message count: %w", err)
	}

	if count >= mm.config.DailyLimit {
		return fmt.Errorf("daily message limit reached (%d/%d)", count, mm.config.DailyLimit)
	}

	logger.Infof("Daily messages: %d/%d", count, mm.config.DailyLimit)
	return nil
}

// findMessageButton finds the Message button on the profile
func (mm *MessageManager) findMessageButton() (*rod.Element, error) {
	// Try different selectors for Message button
	selectors := []string{
		"button[aria-label*='Message']",
		"button:has-text('Message')",
		"div.pvs-profile-actions button:has-text('Message')",
	}

	for _, selector := range selectors {
		element, err := mm.page.Element(selector)
		if err == nil {
			return element, nil
		}
	}

	return nil, fmt.Errorf("message button not found")
}

// typeMessage types the message in the message box
func (mm *MessageManager) typeMessage(message string) error {
	// Wait for message box to appear
	time.Sleep(1 * time.Second)

	// Find message input
	selectors := []string{
		"div.msg-form__contenteditable",
		"div[role='textbox']",
		"div.msg-form__msg-content-container div[contenteditable='true']",
	}

	var messageBox *rod.Element
	var err error

	for _, selector := range selectors {
		messageBox, err = mm.page.Element(selector)
		if err == nil {
			break
		}
	}

	if messageBox == nil {
		return fmt.Errorf("message input not found")
	}

	// Focus and type
	if err := messageBox.Focus(); err != nil {
		return err
	}

	return mm.typer.TypeText(mm.page, messageBox, message)
}

// clickSendButton clicks the Send button
func (mm *MessageManager) clickSendButton() error {
	selectors := []string{
		"button[type='submit']",
		"button.msg-form__send-button",
		"button:has-text('Send')",
	}

	for _, selector := range selectors {
		button, err := mm.page.Element(selector)
		if err == nil {
			return mm.mouse.ClickElement(button)
		}
	}

	return fmt.Errorf("send button not found")
}

// generateMessage generates a personalized message
func (mm *MessageManager) generateMessage(profileName, jobTitle, company string) string {
	if len(mm.config.Templates) == 0 {
		return "Thanks for connecting!"
	}

	// Select random template
	template := mm.config.Templates[mm.rand.Intn(len(mm.config.Templates))]

	// Extract first name
	firstName := strings.Split(profileName, " ")[0]

	// Replace variables
	message := strings.ReplaceAll(template, "{{firstName}}", firstName)
	message = strings.ReplaceAll(message, "{{jobTitle}}", jobTitle)
	message = strings.ReplaceAll(message, "{{company}}", company)

	return message
}

// SendFollowUpMessages sends follow-up messages to newly accepted connections
func (mm *MessageManager) SendFollowUpMessages() error {
	logger.Info("Checking for newly accepted connections")

	// Get uncontacted profiles (this would need to be implemented in the database)
	// For now, we'll skip this functionality
	
	return nil
}
