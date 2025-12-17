package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	
	"github.com/Tanukumar01/linkedin-automation/internal/auth"
	"github.com/Tanukumar01/linkedin-automation/internal/config"
	"github.com/Tanukumar01/linkedin-automation/internal/connections"
	"github.com/Tanukumar01/linkedin-automation/internal/logger"
	"github.com/Tanukumar01/linkedin-automation/internal/messaging"
	"github.com/Tanukumar01/linkedin-automation/internal/search"
	"github.com/Tanukumar01/linkedin-automation/internal/stealth"
	"github.com/Tanukumar01/linkedin-automation/internal/storage"
	"github.com/Tanukumar01/linkedin-automation/pkg/browser"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get config path
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.InitLogger(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting LinkedIn Automation Bot")

	// Load credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		logger.Fatalf("Failed to load credentials: %v", err)
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/linkedin_bot.db"
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		logger.Fatalf("Failed to create data directory: %v", err)
	}

	db, err := storage.NewDB(dbPath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	logger.Info("Database initialized")

	// Initialize browser
	userDataDir := "browser-data"
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		logger.Fatalf("Failed to create browser data directory: %v", err)
	}

	br, err := browser.NewBrowser(cfg.Browser.Headless, userDataDir, cfg.Browser.TimeoutSeconds)
	if err != nil {
		logger.Fatalf("Failed to initialize browser: %v", err)
	}
	defer br.Close()

	logger.Info("Browser initialized")

	// Initialize stealth components
	fingerprint := stealth.NewFingerprintMasker(
		cfg.Browser.UserAgents,
		cfg.Browser.ViewportWidths,
		cfg.Browser.ViewportHeights,
	)

	// Create page with random user agent
	userAgent := fingerprint.GetRandomUserAgent()
	page, err := br.NewPage(userAgent)
	if err != nil {
		logger.Fatalf("Failed to create page: %v", err)
	}

	logger.Infof("Using User-Agent: %s", userAgent)

	// Apply fingerprint masking
	if err := fingerprint.ApplyStealthScripts(page); err != nil {
		logger.Warnf("Failed to apply stealth scripts: %v", err)
	}

	// Randomize viewport
	if err := fingerprint.RandomizeViewport(page); err != nil {
		logger.Warnf("Failed to randomize viewport: %v", err)
	}

	// Initialize stealth controllers
	timing := stealth.NewTimingController(
		cfg.Stealth.Timing.ActionDelayMin,
		cfg.Stealth.Timing.ActionDelayMax,
		cfg.Stealth.Timing.ThinkTimeMin,
		cfg.Stealth.Timing.ThinkTimeMax,
		cfg.Stealth.Timing.ReadingSpeedWPM,
	)

	typer := stealth.NewTyper(
		cfg.Stealth.Typing.WPMMin,
		cfg.Stealth.Typing.WPMMax,
		cfg.Stealth.Typing.TypoProbability,
		cfg.Stealth.Typing.PauseProbability,
	)

	mouse := stealth.NewMouseMover(
		page,
		cfg.Stealth.Mouse.BezierPoints,
		cfg.Stealth.Mouse.SpeedVariation,
		cfg.Stealth.Mouse.OvershootProbability,
		cfg.Stealth.Mouse.MicroCorrectionProbability,
	)

	scroller := stealth.NewScroller(
		cfg.Stealth.Scrolling.SpeedMin,
		cfg.Stealth.Scrolling.SpeedMax,
		cfg.Stealth.Scrolling.ScrollBackProbability,
		cfg.Stealth.Scrolling.PauseProbability,
	)

	scheduler, err := stealth.NewScheduler(
		cfg.Stealth.Scheduling.BusinessHoursStart,
		cfg.Stealth.Scheduling.BusinessHoursEnd,
		cfg.Stealth.Scheduling.Timezone,
		cfg.Stealth.Scheduling.WeekendActivity,
		cfg.Stealth.Scheduling.BreakDurationMin,
		cfg.Stealth.Scheduling.BreakDurationMax,
		cfg.Stealth.Scheduling.BreakProbability,
	)
	if err != nil {
		logger.Fatalf("Failed to initialize scheduler: %v", err)
	}

	logger.Info("Stealth components initialized")

	// Check if within business hours
	if !scheduler.IsBusinessHours() {
		logger.Info("Outside business hours, waiting...")
		scheduler.WaitForBusinessHours()
	}

	// Initialize authentication
	authenticator := auth.NewAuthenticator(page, typer, timing, "cookies.json")

	// Login
	logger.Info("Attempting to login...")
	if err := authenticator.Login(creds.Email, creds.Password); err != nil {
		logger.Fatalf("Login failed: %v", err)
	}

	logger.Info("Successfully logged in")

	// Log activity
	db.LogActivity("login", "Successful login")

	// Initialize search
	searcher := search.NewSearcher(page, &cfg.Search, db, timing, scroller)

	// Initialize connection manager
	connManager := connections.NewConnectionManager(page, &cfg.Connections, db, timing, typer, mouse, scroller)

	// Initialize message manager
	msgManager := messaging.NewMessageManager(page, &cfg.Messaging, db, timing, typer, mouse, scroller)

	// Suppress unused variable warning
	_ = msgManager

	// Main automation loop
	logger.Info("Starting automation workflow")

	// Step 1: Search for profiles
	logger.Info("Searching for profiles...")
	results, err := searcher.Search()
	if err != nil {
		logger.Errorf("Search failed: %v", err)
	} else {
		logger.Infof("Found %d profiles", len(results))
	}

	// Step 2: Send connection requests
	logger.Info("Sending connection requests...")
	uncontactedProfiles, err := db.GetUncontactedProfiles(cfg.Connections.DailyLimit)
	if err != nil {
		logger.Errorf("Failed to get uncontacted profiles: %v", err)
	} else {
		for _, profile := range uncontactedProfiles {
			// Check if should take a break
			if scheduler.ShouldTakeBreak() {
				logger.Info("Taking a break...")
				scheduler.TakeBreak()
			}

			if err := connManager.SendConnectionRequest(profile.ProfileURL, profile.ProfileName, profile.JobTitle, profile.Company); err != nil {
				logger.Errorf("Failed to send connection request: %v", err)
				
				// Check if daily limit reached
				if err.Error() == fmt.Sprintf("daily connection limit reached (%d/%d)", cfg.Connections.DailyLimit, cfg.Connections.DailyLimit) {
					logger.Info("Daily connection limit reached, stopping")
					break
				}
			}
		}
	}

	// Step 3: Send follow-up messages (optional)
	// This would require detecting newly accepted connections
	// For now, we'll skip this step

	logger.Info("Automation workflow completed")

	// Print daily stats
	stats, err := db.GetDailyStats(time.Now())
	if err == nil {
		logger.Infof("Daily Stats:")
		logger.Infof("  Connections Sent: %d", stats.ConnectionsSent)
		logger.Infof("  Connections Accepted: %d", stats.ConnectionsAccepted)
		logger.Infof("  Messages Sent: %d", stats.MessagesSent)
		logger.Infof("  Searches Performed: %d", stats.SearchesPerformed)
	}

	logger.Info("LinkedIn Automation Bot finished")
}
