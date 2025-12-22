package browser

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Browser wraps Rod browser with additional functionality
type Browser struct {
	browser *rod.Browser
	page    *rod.Page
	timeout time.Duration
}

// NewBrowser creates a new browser instance
func NewBrowser(headless bool, userDataDir string, timeoutSeconds int) (*Browser, error) {
	// Launch browser
	l := launcher.New().
		Headless(headless).
		UserDataDir(userDataDir).
		Leakless(false).
		NoSandbox(true).
		Set("disable-gpu")

	// Print browser info for debugging
	if path, exists := launcher.LookPath(); exists {
		fmt.Printf("Launching browser: %s\n", path)
		l.Bin(path)
	}

	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	fmt.Printf("Browser launched! Debug URL: %s\n", url)

	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

	return &Browser{
		browser: browser,
		timeout: timeout,
	}, nil
}

// NewPage creates a new page with stealth settings
func (b *Browser) NewPage(userAgent string) (*rod.Page, error) {
	page, err := stealth.Page(b.browser)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set user agent
	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: userAgent,
	}); err != nil {
		return nil, fmt.Errorf("failed to set user agent: %w", err)
	}

	// Set timeout (disabled globally to avoid 'context deadline exceeded' on the whole page)
	// page = page.Timeout(b.timeout)

	b.page = page
	return page, nil
}

// GetPage returns the current page
func (b *Browser) GetPage() *rod.Page {
	return b.page
}

// Navigate navigates to a URL
func (b *Browser) Navigate(url string) error {
	if b.page == nil {
		return fmt.Errorf("no page available")
	}

	return b.page.Navigate(url)
}

// WaitLoad waits for page to load
func (b *Browser) WaitLoad() error {
	if b.page == nil {
		return fmt.Errorf("no page available")
	}

	return b.page.WaitLoad()
}

// Screenshot takes a screenshot
func (b *Browser) Screenshot(path string) error {
	if b.page == nil {
		return fmt.Errorf("no page available")
	}

	data, err := b.page.Screenshot(true, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Close closes the browser
func (b *Browser) Close() error {
	if b.page != nil {
		b.page.Close()
	}
	if b.browser != nil {
		return b.browser.Close()
	}
	return nil
}

// WaitForElement waits for an element to appear
func (b *Browser) WaitForElement(selector string) (*rod.Element, error) {
	if b.page == nil {
		return nil, fmt.Errorf("no page available")
	}

	return b.page.Element(selector)
}

// WaitForElements waits for elements to appear
func (b *Browser) WaitForElements(selector string) (rod.Elements, error) {
	if b.page == nil {
		return nil, fmt.Errorf("no page available")
	}

	return b.page.Elements(selector)
}

// HasElement checks if an element exists
func (b *Browser) HasElement(selector string) bool {
	if b.page == nil {
		return false
	}

	has, _, err := b.page.Has(selector)
	return err == nil && has
}

// GetText gets text from an element
func (b *Browser) GetText(selector string) (string, error) {
	element, err := b.WaitForElement(selector)
	if err != nil {
		return "", err
	}

	return element.Text()
}

// Click clicks an element
func (b *Browser) Click(selector string) error {
	element, err := b.WaitForElement(selector)
	if err != nil {
		return err
	}

	element.MustClick()
	return nil
}

// Type types text into an element
func (b *Browser) Type(selector, text string) error {
	element, err := b.WaitForElement(selector)
	if err != nil {
		return err
	}

	return element.Input(text)
}

// GetCurrentURL returns the current page URL
func (b *Browser) GetCurrentURL() string {
	if b.page == nil {
		return ""
	}

	info := b.page.MustInfo()
	return info.URL
}
