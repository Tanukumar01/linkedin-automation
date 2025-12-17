package auth

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	
	"github.com/Tanukumar01/linkedin-automation/internal/logger"
	"github.com/Tanukumar01/linkedin-automation/internal/stealth"
)

// Authenticator handles LinkedIn authentication
type Authenticator struct {
	page          *rod.Page
	typer         *stealth.Typer
	timing        *stealth.TimingController
	cookieManager *CookieManager
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(page *rod.Page, typer *stealth.Typer, timing *stealth.TimingController, cookieFile string) *Authenticator {
	return &Authenticator{
		page:          page,
		typer:         typer,
		timing:        timing,
		cookieManager: NewCookieManager(cookieFile),
	}
}

// Login performs LinkedIn login
func (a *Authenticator) Login(email, password string) error {
	logger.Info("Starting LinkedIn login process")

	// Try to load existing cookies
	if err := a.cookieManager.LoadCookies(a.page); err != nil {
		logger.Warnf("Failed to load cookies: %v", err)
	}

	// Navigate to LinkedIn
	if err := a.page.Navigate("https://www.linkedin.com"); err != nil {
		return fmt.Errorf("failed to navigate to LinkedIn: %w", err)
	}

	if err := a.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	a.timing.Wait(a.timing.ThinkTime())

	// Check if already logged in
	if a.IsLoggedIn() {
		logger.Info("Already logged in using saved session")
		return nil
	}

	logger.Info("No valid session found, performing login")

	// Navigate to login page
	if err := a.page.Navigate("https://www.linkedin.com/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	if err := a.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for login page: %w", err)
	}

	a.timing.Wait(a.timing.ThinkTime())

	// Find email input
	emailInput, err := a.page.Element("#username")
	if err != nil {
		return fmt.Errorf("failed to find email input: %w", err)
	}

	// Type email
	logger.Info("Entering email")
	if err := a.typer.TypeText(a.page, emailInput, email); err != nil {
		return fmt.Errorf("failed to type email: %w", err)
	}

	a.timing.Wait(a.timing.ShortPause())

	// Find password input
	passwordInput, err := a.page.Element("#password")
	if err != nil {
		return fmt.Errorf("failed to find password input: %w", err)
	}

	// Type password
	logger.Info("Entering password")
	if err := a.typer.TypeText(a.page, passwordInput, password); err != nil {
		return fmt.Errorf("failed to type password: %w", err)
	}

	a.timing.Wait(a.timing.ThinkTime())

	// Click sign in button
	logger.Info("Clicking sign in button")
	signInButton, err := a.page.Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("failed to find sign in button: %w", err)
	}

	if err := signInButton.Click(rod.ButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click sign in button: %w", err)
	}

	// Wait for navigation
	time.Sleep(5 * time.Second)

	// Check for security challenges
	if err := a.checkForSecurityChallenges(); err != nil {
		return err
	}

	// Verify login success
	if !a.IsLoggedIn() {
		return fmt.Errorf("login failed - not logged in after authentication")
	}

	logger.Info("Login successful")

	// Save cookies
	if err := a.cookieManager.SaveCookies(a.page); err != nil {
		logger.Warnf("Failed to save cookies: %v", err)
	}

	return nil
}

// IsLoggedIn checks if user is logged in
func (a *Authenticator) IsLoggedIn() bool {
	// Check for feed or profile elements that indicate logged-in state
	has, _, err := a.page.Has("nav.global-nav")
	if err != nil {
		return false
	}

	return has
}

// checkForSecurityChallenges detects security challenges
func (a *Authenticator) checkForSecurityChallenges() error {
	// Check for 2FA
	has2FA, _, _ := a.page.Has("input[id*='verification']")
	if has2FA {
		logger.Warn("2FA detected - manual intervention required")
		return fmt.Errorf("2FA challenge detected - please complete manually")
	}

	// Check for CAPTCHA
	hasCaptcha, _, _ := a.page.Has("iframe[title*='recaptcha']")
	if hasCaptcha {
		logger.Warn("CAPTCHA detected - manual intervention required")
		return fmt.Errorf("CAPTCHA challenge detected - please complete manually")
	}

	// Check for unusual login alert
	hasAlert, _, _ := a.page.Has("div[data-test-id='unusual-activity']")
	if hasAlert {
		logger.Warn("Unusual login activity alert detected")
		return fmt.Errorf("unusual login activity detected - please verify manually")
	}

	// Check for email verification
	hasEmailVerification, _, _ := a.page.Has("input[name='pin']")
	if hasEmailVerification {
		logger.Warn("Email verification required - manual intervention needed")
		return fmt.Errorf("email verification required - please complete manually")
	}

	return nil
}

// Logout performs logout
func (a *Authenticator) Logout() error {
	logger.Info("Logging out")

	// Navigate to logout URL
	if err := a.page.Navigate("https://www.linkedin.com/m/logout"); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	time.Sleep(2 * time.Second)

	// Clear cookies
	if err := a.cookieManager.ClearCookies(); err != nil {
		logger.Warnf("Failed to clear cookies: %v", err)
	}

	return nil
}

// GetCookieManager returns the cookie manager
func (a *Authenticator) GetCookieManager() *CookieManager {
	return a.cookieManager
}
