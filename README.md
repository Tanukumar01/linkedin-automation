# LinkedIn Automation Tool

A production-grade LinkedIn automation system built with **Go** and **Rod browser automation**, featuring sophisticated anti-detection mechanisms and clean, modular architecture.

## ⚠️ Disclaimer

This tool is for **educational purposes only**. Automated interaction with LinkedIn may violate their Terms of Service. Use at your own risk. The authors are not responsible for any account restrictions or bans resulting from the use of this tool.

## ✨ Features

### Core Functionality
- **Authentication System**: Secure login with session cookie persistence
- **Smart Search**: Search LinkedIn users by job title, company, location, and keywords
- **Connection Automation**: Send personalized connection requests with custom notes
- **Messaging System**: Automated follow-up messages with template support
- **State Persistence**: SQLite database tracks all activities and prevents duplicates

### Anti-Detection Mechanisms (10+ Techniques)

#### Mandatory Techniques
1. **Bézier Curve Mouse Movement**: Natural cursor paths with overshoot and micro-corrections
2. **Randomized Timing Patterns**: Variable delays, think time, and reading simulation
3. **Browser Fingerprint Masking**: Disables `navigator.webdriver`, randomizes viewport, masks automation properties

#### Additional Techniques
4. **Realistic Typing Simulation**: Variable speed, typos, corrections, natural pauses
5. **Natural Scrolling**: Acceleration, deceleration, scroll-back, random pauses
6. **Activity Scheduling**: Business hours operation, weekend detection, random breaks
7. **Rate Limiting**: Daily/hourly limits, cooldown periods, exponential backoff
8. **Mouse Hovering**: Idle cursor wandering, element hovering
9. **Viewport Randomization**: Dynamic viewport dimensions
10. **Human Reading Patterns**: Content-based reading time simulation

## 📋 Prerequisites

- **Go 1.21+** installed
- **Chrome/Chromium** browser installed
- LinkedIn account credentials

## 🚀 Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/Tanukumar01/linkedin-automation.git
   cd linkedin-automation
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Configure environment variables**:
   ```bash
   cp .env.example .env
   ```

   Edit `.env` and add your credentials:
   ```env
   LINKEDIN_EMAIL=your-email@example.com
   LINKEDIN_PASSWORD=your-password
   LOG_LEVEL=info
   HEADLESS_MODE=false
   ```

4. **Customize configuration** (optional):
   Edit `configs/config.yaml` to adjust:
   - Search filters
   - Connection/message limits
   - Stealth settings
   - Message templates

## 🎯 Usage

### Run the bot:
```bash
go run main.go
```

### Build executable:
```bash
go build -o linkedin-bot main.go
./linkedin-bot
```

### Configuration Options

#### Search Filters (`configs/config.yaml`)
```yaml
search:
  filters:
    job_titles:
      - "Software Engineer"
      - "Senior Developer"
    companies:
      - "Google"
      - "Microsoft"
    locations:
      - "United States"
    keywords:
      - "golang"
      - "backend"
```

#### Connection Settings
```yaml
connections:
  daily_limit: 20          # Max connections per day
  hourly_limit: 5          # Max connections per hour
  note_templates:
    - "Hi {{firstName}}, I came across your profile..."
```

#### Stealth Settings
```yaml
stealth:
  timing:
    action_delay_min: 2    # Minimum delay between actions (seconds)
    action_delay_max: 5    # Maximum delay between actions (seconds)
  
  scheduling:
    business_hours_start: 9
    business_hours_end: 18
    timezone: "America/New_York"
```

## 📁 Project Structure

```
linkedin-automation/
├── cmd/linkedin-bot/       # Application entry point
├── internal/
│   ├── auth/              # Authentication & session management
│   ├── search/            # Search & profile discovery
│   ├── connections/       # Connection request management
│   ├── messaging/         # Messaging system
│   ├── stealth/           # Anti-detection mechanisms
│   ├── config/            # Configuration management
│   ├── storage/           # Database & state persistence
│   └── logger/            # Structured logging
├── pkg/browser/           # Browser automation wrapper
├── configs/               # Configuration files
└── README.md
```

## 🛡️ Anti-Detection Strategy

### How It Works

1. **Human-like Mouse Movement**
   - Uses Bézier curves for natural cursor paths
   - Random overshoot and micro-corrections
   - Variable speed with acceleration/deceleration

2. **Timing Randomization**
   - Random delays between all actions
   - Think time before interactions
   - Reading time based on content length

3. **Browser Fingerprinting**
   - Masks `navigator.webdriver` property
   - Randomizes User-Agent strings
   - Injects realistic browser properties

4. **Activity Patterns**
   - Operates only during business hours
   - Takes random breaks
   - Varies daily activity timing

5. **Rate Limiting**
   - Conservative daily limits (default: 20 connections/day)
   - Cooldown periods between actions
   - Exponential backoff on errors

## 📊 Database Schema

The tool uses SQLite to track:
- **Connection Requests**: Profile URL, name, status, timestamps
- **Messages**: Sent messages with content and timestamps
- **Search Results**: Cached profiles with metadata
- **Activity Logs**: All actions for auditing

## 🔧 Troubleshooting

### Common Issues

**Go not found**:
```bash
# Install Go from https://golang.org/dl/
# Verify installation:
go version
```

**Browser not found**:
- Ensure Chrome/Chromium is installed
- Rod will attempt to download Chromium if not found

**Login fails**:
- Verify credentials in `.env`
- Check for 2FA/CAPTCHA (manual intervention required)
- Review logs for specific error messages

**Daily limit reached**:
- Adjust `daily_limit` in `configs/config.yaml`
- Wait 24 hours for limit reset

## 📝 Logging

Logs are output to stdout with configurable levels:
- `debug`: Detailed debugging information
- `info`: General information (default)
- `warn`: Warning messages
- `error`: Error messages

Set log level in `.env`:
```env
LOG_LEVEL=debug
```

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## ⚖️ Legal Notice

This tool automates interactions with LinkedIn, which may violate their Terms of Service. Use responsibly and at your own risk. The authors assume no liability for:
- Account restrictions or bans
- Data loss
- Any other consequences of using this tool

**Always review and comply with LinkedIn's Terms of Service and Automation Policy.**

## 🙏 Acknowledgments

- [Rod](https://github.com/go-rod/rod) - Browser automation library
- [Rod Stealth](https://github.com/go-rod/stealth) - Stealth plugin
- Go community for excellent tooling

---

**Built with ❤️ using Go and Rod**
