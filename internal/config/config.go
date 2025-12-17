package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Search      SearchConfig      `yaml:"search"`
	Connections ConnectionsConfig `yaml:"connections"`
	Messaging   MessagingConfig   `yaml:"messaging"`
	Stealth     StealthConfig     `yaml:"stealth"`
	Browser     BrowserConfig     `yaml:"browser"`
	Logging     LoggingConfig     `yaml:"logging"`
}

// SearchConfig contains search-related settings
type SearchConfig struct {
	MaxResults         int      `yaml:"max_results"`
	PaginationDelayMin int      `yaml:"pagination_delay_min"`
	PaginationDelayMax int      `yaml:"pagination_delay_max"`
	Filters            Filters  `yaml:"filters"`
}

// Filters contains search filter criteria
type Filters struct {
	JobTitles []string `yaml:"job_titles"`
	Companies []string `yaml:"companies"`
	Locations []string `yaml:"locations"`
	Keywords  []string `yaml:"keywords"`
}

// ConnectionsConfig contains connection request settings
type ConnectionsConfig struct {
	DailyLimit                  int      `yaml:"daily_limit"`
	HourlyLimit                 int      `yaml:"hourly_limit"`
	NoteTemplates               []string `yaml:"note_templates"`
	NoteCharacterLimit          int      `yaml:"note_character_limit"`
	CooldownBetweenRequestsMin  int      `yaml:"cooldown_between_requests_min"`
	CooldownBetweenRequestsMax  int      `yaml:"cooldown_between_requests_max"`
}

// MessagingConfig contains messaging settings
type MessagingConfig struct {
	DailyLimit                 int      `yaml:"daily_limit"`
	HourlyLimit                int      `yaml:"hourly_limit"`
	Templates                  []string `yaml:"templates"`
	CooldownBetweenMessagesMin int      `yaml:"cooldown_between_messages_min"`
	CooldownBetweenMessagesMax int      `yaml:"cooldown_between_messages_max"`
}

// StealthConfig contains anti-detection settings
type StealthConfig struct {
	Mouse      MouseConfig      `yaml:"mouse"`
	Timing     TimingConfig     `yaml:"timing"`
	Typing     TypingConfig     `yaml:"typing"`
	Scrolling  ScrollingConfig  `yaml:"scrolling"`
	Scheduling SchedulingConfig `yaml:"scheduling"`
}

// MouseConfig contains mouse movement settings
type MouseConfig struct {
	BezierPoints              int     `yaml:"bezier_points"`
	SpeedVariation            float64 `yaml:"speed_variation"`
	OvershootProbability      float64 `yaml:"overshoot_probability"`
	MicroCorrectionProbability float64 `yaml:"micro_correction_probability"`
}

// TimingConfig contains timing-related settings
type TimingConfig struct {
	ActionDelayMin  int `yaml:"action_delay_min"`
	ActionDelayMax  int `yaml:"action_delay_max"`
	ThinkTimeMin    int `yaml:"think_time_min"`
	ThinkTimeMax    int `yaml:"think_time_max"`
	ReadingSpeedWPM int `yaml:"reading_speed_wpm"`
}

// TypingConfig contains typing simulation settings
type TypingConfig struct {
	WPMMin           int     `yaml:"wpm_min"`
	WPMMax           int     `yaml:"wpm_max"`
	TypoProbability  float64 `yaml:"typo_probability"`
	PauseProbability float64 `yaml:"pause_probability"`
}

// ScrollingConfig contains scrolling behavior settings
type ScrollingConfig struct {
	SpeedMin              int     `yaml:"speed_min"`
	SpeedMax              int     `yaml:"speed_max"`
	ScrollBackProbability float64 `yaml:"scroll_back_probability"`
	PauseProbability      float64 `yaml:"pause_probability"`
}

// SchedulingConfig contains activity scheduling settings
type SchedulingConfig struct {
	BusinessHoursStart int     `yaml:"business_hours_start"`
	BusinessHoursEnd   int     `yaml:"business_hours_end"`
	Timezone           string  `yaml:"timezone"`
	WeekendActivity    bool    `yaml:"weekend_activity"`
	BreakDurationMin   int     `yaml:"break_duration_min"`
	BreakDurationMax   int     `yaml:"break_duration_max"`
	BreakProbability   float64 `yaml:"break_probability"`
}

// BrowserConfig contains browser settings
type BrowserConfig struct {
	Headless        bool     `yaml:"headless"`
	UserAgents      []string `yaml:"user_agents"`
	ViewportWidths  []int    `yaml:"viewport_widths"`
	ViewportHeights []int    `yaml:"viewport_heights"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Credentials contains LinkedIn login credentials
type Credentials struct {
	Email    string
	Password string
}

// LoadConfig loads configuration from YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	if headless := os.Getenv("HEADLESS_MODE"); headless == "true" {
		config.Browser.Headless = true
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadCredentials loads LinkedIn credentials from environment variables
func LoadCredentials() (*Credentials, error) {
	email := os.Getenv("LINKEDIN_EMAIL")
	password := os.Getenv("LINKEDIN_PASSWORD")

	if email == "" || password == "" {
		return nil, fmt.Errorf("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set in environment variables")
	}

	return &Credentials{
		Email:    email,
		Password: password,
	}, nil
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	if config.Search.MaxResults <= 0 {
		return fmt.Errorf("search.max_results must be greater than 0")
	}

	if config.Connections.DailyLimit <= 0 {
		return fmt.Errorf("connections.daily_limit must be greater than 0")
	}

	if config.Messaging.DailyLimit <= 0 {
		return fmt.Errorf("messaging.daily_limit must be greater than 0")
	}

	if config.Browser.TimeoutSeconds <= 0 {
		return fmt.Errorf("browser.timeout_seconds must be greater than 0")
	}

	if len(config.Browser.UserAgents) == 0 {
		return fmt.Errorf("browser.user_agents must contain at least one user agent")
	}

	// Validate timezone
	if _, err := time.LoadLocation(config.Stealth.Scheduling.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}
