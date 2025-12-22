package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// CookieManager handles cookie persistence
type CookieManager struct {
	cookieFile string
}

// NewCookieManager creates a new cookie manager
func NewCookieManager(cookieFile string) *CookieManager {
	return &CookieManager{
		cookieFile: cookieFile,
	}
}

// SaveCookies saves cookies to file
func (cm *CookieManager) SaveCookies(page *rod.Page) error {
	cookies, err := page.Cookies([]string{})
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cookies: %w", err)
	}

	if err := os.WriteFile(cm.cookieFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cookies file: %w", err)
	}

	return nil
}

// LoadCookies loads cookies from file
func (cm *CookieManager) LoadCookies(page *rod.Page) error {
	// Check if cookie file exists
	if _, err := os.Stat(cm.cookieFile); os.IsNotExist(err) {
		return nil // No cookies to load
	}

	data, err := os.ReadFile(cm.cookieFile)
	if err != nil {
		return fmt.Errorf("failed to read cookies file: %w", err)
	}

	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	var params []*proto.NetworkCookieParam
	for _, c := range cookies {
		params = append(params, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
			Expires:  c.Expires,
		})
	}

	// Set cookies
	if err := page.SetCookies(params); err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}

	return nil
}

// ClearCookies removes the cookie file
func (cm *CookieManager) ClearCookies() error {
	if _, err := os.Stat(cm.cookieFile); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(cm.cookieFile)
}

// AreCookiesValid checks if cookies are still valid
func (cm *CookieManager) AreCookiesValid(page *rod.Page) bool {
	cookies, err := page.Cookies([]string{})
	if err != nil {
		return false
	}

	// Check if we have LinkedIn session cookies
	hasSessionCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "li_at" || cookie.Name == "JSESSIONID" {
			// Check if cookie is not expired
			if cookie.Expires > 0 && time.Unix(int64(cookie.Expires), 0).After(time.Now()) {
				hasSessionCookie = true
				break
			}
		}
	}

	return hasSessionCookie
}
