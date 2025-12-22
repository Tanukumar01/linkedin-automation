package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// FingerprintMasker handles browser fingerprint masking
type FingerprintMasker struct {
	userAgents      []string
	viewportWidths  []int
	viewportHeights []int
	rand            *rand.Rand
}

// NewFingerprintMasker creates a new fingerprint masker
func NewFingerprintMasker(userAgents []string, viewportWidths, viewportHeights []int) *FingerprintMasker {
	return &FingerprintMasker{
		userAgents:      userAgents,
		viewportWidths:  viewportWidths,
		viewportHeights: viewportHeights,
		rand:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomUserAgent returns a random user agent
func (f *FingerprintMasker) GetRandomUserAgent() string {
	if len(f.userAgents) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	return f.userAgents[f.rand.Intn(len(f.userAgents))]
}

// GetRandomViewport returns random viewport dimensions
func (f *FingerprintMasker) GetRandomViewport() (int, int) {
	width := f.viewportWidths[f.rand.Intn(len(f.viewportWidths))]
	height := f.viewportHeights[f.rand.Intn(len(f.viewportHeights))]
	return width, height
}

// ApplyStealthScripts applies stealth scripts to mask automation
func (f *FingerprintMasker) ApplyStealthScripts(page *rod.Page) error {
	// Disable navigator.webdriver
	_, err := page.Eval(`() => {
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});
	}`)
	if err != nil {
		return err
	}

	// Mask chrome automation properties
	_, err = page.Eval(`() => {
		window.navigator.chrome = {
			runtime: {},
		};
	}`)
	if err != nil {
		return err
	}

	// Override permissions
	_, err = page.Eval(`() => {
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	}`)
	if err != nil {
		return err
	}

	// Mock plugins
	_, err = page.Eval(`() => {
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				},
				{
					0: {type: "application/pdf", suffixes: "pdf", description: ""},
					description: "",
					filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
					length: 1,
					name: "Chrome PDF Viewer"
				}
			],
		});
	}`)
	if err != nil {
		return err
	}

	// Mock languages
	_, err = page.Eval(`() => {
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en'],
		});
	}`)
	if err != nil {
		return err
	}

	// Override toString methods to hide modifications
	_, err = page.Eval(`() => {
		const originalToString = Function.prototype.toString;
		Function.prototype.toString = function() {
			if (this === navigator.permissions.query) {
				return 'function query() { [native code] }';
			}
			return originalToString.call(this);
		};
	}`)
	if err != nil {
		return err
	}

	return nil
}

// RandomizeViewport randomly changes the viewport size
func (f *FingerprintMasker) RandomizeViewport(page *rod.Page) error {
	width, height := f.GetRandomViewport()
	return page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 1,
		Mobile:            false,
	})
}
