# LinkedIn Automation Tool - Setup & Testing Guide

## 📋 Prerequisites Checklist

Before running this project, ensure you have:
- [ ] Go 1.21 or higher installed
- [ ] Chrome or Chromium browser installed
- [ ] LinkedIn account credentials
- [ ] Text editor (VS Code recommended)

---

## 🔧 Step 1: Install Go

### Windows Installation

1. **Download Go**:
   - Visit: https://go.dev/dl/
   - Download the Windows installer (e.g., `go1.21.windows-amd64.msi`)

2. **Run the installer**:
   - Double-click the downloaded `.msi` file
   - Follow the installation wizard
   - Default installation path: `C:\Program Files\Go`

3. **Verify installation**:
   ```powershell
   go version
   ```
   Expected output: `go version go1.21.x windows/amd64`

4. **Check Go environment**:
   ```powershell
   go env GOPATH
   go env GOROOT
   ```

### If Go is not recognized:
Add Go to your PATH:
1. Open System Properties → Environment Variables
2. Under "System variables", find `Path`
3. Add: `C:\Program Files\Go\bin`
4. Restart PowerShell/Terminal

---

## 📦 Step 2: Install Dependencies

Navigate to the project directory and install all required packages:

```powershell
cd c:\Users\tanuk\OneDrive\LinkedIn_Automation_tool

# Download all dependencies
go mod download

# Verify dependencies
go mod verify
```

### Expected Dependencies:
- `github.com/go-rod/rod` - Browser automation
- `github.com/go-rod/stealth` - Stealth plugin
- `github.com/joho/godotenv` - Environment variables
- `github.com/mattn/go-sqlite3` - SQLite database
- `go.uber.org/zap` - Structured logging
- `gopkg.in/yaml.v3` - YAML parsing

---

## ⚙️ Step 3: Configure the Application

### 3.1 Create Environment File

```powershell
# Copy the example file
Copy-Item .env.example .env

# Edit the .env file with your credentials
notepad .env
```

### 3.2 Add Your LinkedIn Credentials

Edit `.env` and replace with your actual credentials:

```env
# LinkedIn Credentials (REQUIRED)
LINKEDIN_EMAIL=your-actual-email@example.com
LINKEDIN_PASSWORD=your-actual-password

# Application Settings
LOG_LEVEL=info
HEADLESS_MODE=false
CONFIG_PATH=configs/config.yaml

# Database
DB_PATH=data/linkedin_bot.db

# Browser Settings
BROWSER_TIMEOUT=30
```

⚠️ **Security Note**: Never commit the `.env` file to git!

### 3.3 Customize Configuration (Optional)

Edit `configs/config.yaml` to customize:

```yaml
# Example: Change search filters
search:
  filters:
    job_titles:
      - "Your Target Job Title"
      - "Another Job Title"
    companies:
      - "Target Company 1"
      - "Target Company 2"
    locations:
      - "Your Target Location"

# Example: Adjust daily limits
connections:
  daily_limit: 10  # Start conservative
  
messaging:
  daily_limit: 5   # Start conservative
```

---

## 🚀 Step 4: Run the Application

### Option A: Run Directly (Development)

```powershell
# Run with visible browser (recommended for first run)
go run main.go
```

### Option B: Build and Run (Production)

```powershell
# Build the executable
go build -o linkedin-bot.exe main.go

# Run the executable
.\linkedin-bot.exe
```

### What to Expect:

1. **Initialization**:
   ```
   Starting LinkedIn Automation Bot
   Database initialized
   Browser initialized
   Using User-Agent: Mozilla/5.0...
   Stealth components initialized
   ```

2. **Business Hours Check**:
   ```
   Outside business hours, waiting...
   # OR
   Within business hours, proceeding...
   ```

3. **Authentication**:
   ```
   Attempting to login...
   Entering email
   Entering password
   Clicking sign in button
   Login successful
   ```

4. **Automation Workflow**:
   ```
   Starting automation workflow
   Searching for profiles...
   Found 25 profiles
   Sending connection requests...
   Daily connections: 5/20
   ```

5. **Completion**:
   ```
   Automation workflow completed
   Daily Stats:
     Connections Sent: 5
     Messages Sent: 0
   LinkedIn Automation Bot finished
   ```

---

## 🧪 Step 5: Testing Strategy

### 5.1 Initial Test Run (Safe Mode)

**First time? Start with these safe settings:**

Edit `configs/config.yaml`:
```yaml
connections:
  daily_limit: 3  # Very conservative
  
search:
  max_results: 10  # Small batch
  
browser:
  headless: false  # Watch what happens
```

**Run and observe**:
```powershell
go run main.go
```

Watch the browser window to see:
- ✅ Natural mouse movements
- ✅ Realistic typing with pauses
- ✅ Smooth scrolling
- ✅ Random delays between actions

### 5.2 Test Individual Components

#### Test Authentication Only:
Comment out the automation workflow in `main.go` (lines 200-240):
```go
// Step 1: Search for profiles
// logger.Info("Searching for profiles...")
// ...
```

Then run:
```powershell
go run main.go
```

Should stop after "Successfully logged in"

#### Test Search Only:
Keep only the search section uncommented:
```go
// Step 1: Search for profiles
logger.Info("Searching for profiles...")
results, err := searcher.Search()
// ...

// Comment out Step 2 and Step 3
```

#### Test Database:
Check if database was created:
```powershell
# Database should exist
Test-Path data\linkedin_bot.db
```

### 5.3 Verify Stealth Mechanisms

**Watch for these behaviors** (non-headless mode):

1. **Mouse Movement**:
   - ✅ Cursor follows curved paths (not straight lines)
   - ✅ Occasional overshoot and correction
   - ✅ Variable speed

2. **Typing**:
   - ✅ Variable typing speed
   - ✅ Occasional typos with backspace
   - ✅ Pauses between words

3. **Scrolling**:
   - ✅ Smooth acceleration/deceleration
   - ✅ Occasional scroll-back
   - ✅ Random pauses

4. **Timing**:
   - ✅ Random delays between actions (2-5 seconds)
   - ✅ Think time before clicks (1-3 seconds)

### 5.4 Check Logs

Logs show all activities:
```powershell
# Run with debug logging
$env:LOG_LEVEL="debug"
go run main.go
```

Look for:
- `[INFO]` - General information
- `[WARN]` - Warnings (non-critical)
- `[ERROR]` - Errors (investigate these)
- `[DEBUG]` - Detailed debugging info

### 5.5 Inspect Database

Use SQLite browser or command line:
```powershell
# Install sqlite3 (if not installed)
# Download from: https://www.sqlite.org/download.html

# View database
sqlite3 data\linkedin_bot.db

# Run queries
.tables
SELECT * FROM connection_requests;
SELECT * FROM activity_logs;
.exit
```

---

## 🐛 Troubleshooting

### Issue: "go: command not found"
**Solution**: Install Go (see Step 1)

### Issue: "failed to launch browser"
**Solution**: 
- Ensure Chrome/Chromium is installed
- Rod will auto-download Chromium if needed
- Check antivirus isn't blocking

### Issue: "Login failed"
**Solution**:
- Verify credentials in `.env`
- Check for 2FA (requires manual completion)
- Look for CAPTCHA (requires manual completion)
- Review logs for specific error

### Issue: "CAPTCHA detected"
**Solution**:
- Complete CAPTCHA manually in the browser window
- The bot will wait for you
- Consider using headless=false for first runs

### Issue: "Daily limit reached"
**Solution**:
- Wait 24 hours for reset
- Or adjust `daily_limit` in config.yaml

### Issue: "Element not found"
**Solution**:
- LinkedIn may have changed their UI
- Update selectors in the relevant files:
  - `internal/auth/login.go` - Login selectors
  - `internal/search/search.go` - Search selectors
  - `internal/connections/connections.go` - Connection selectors

### Issue: "Database locked"
**Solution**:
- Close any other processes using the database
- Delete `data/linkedin_bot.db` to start fresh

---

## 📊 Monitoring & Analytics

### View Daily Statistics

The bot prints stats at the end:
```
Daily Stats:
  Connections Sent: 15
  Connections Accepted: 3
  Messages Sent: 5
  Searches Performed: 2
```

### Query Database for Insights

```sql
-- View all connection requests
SELECT profile_name, job_title, company, status, sent_at 
FROM connection_requests 
ORDER BY sent_at DESC;

-- View acceptance rate
SELECT 
  COUNT(*) as total,
  SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END) as accepted,
  ROUND(100.0 * SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END) / COUNT(*), 2) as acceptance_rate
FROM connection_requests;

-- View activity timeline
SELECT action, COUNT(*) as count, DATE(timestamp) as date
FROM activity_logs
GROUP BY action, DATE(timestamp)
ORDER BY date DESC;
```

---

## 🔒 Safety Best Practices

### Start Slow
1. **Day 1**: 3-5 connections, observe
2. **Day 2-3**: 5-10 connections, monitor for issues
3. **Day 4+**: Gradually increase to 15-20 (max)

### Monitor Account Health
- Check LinkedIn notifications for warnings
- Watch for unusual activity alerts
- Stop immediately if you see security challenges

### Use Conservative Settings
```yaml
connections:
  daily_limit: 15  # Don't exceed 20-25
  hourly_limit: 5  # Spread throughout day

stealth:
  scheduling:
    business_hours_start: 9
    business_hours_end: 17  # Don't run 24/7
    weekend_activity: false  # Rest on weekends
```

### Regular Breaks
```yaml
stealth:
  scheduling:
    break_probability: 0.2  # 20% chance of break
    break_duration_min: 30  # 30-90 minute breaks
    break_duration_max: 90
```

---

## 🎯 Quick Start Checklist

- [ ] Install Go 1.21+
- [ ] Run `go mod download`
- [ ] Copy `.env.example` to `.env`
- [ ] Add LinkedIn credentials to `.env`
- [ ] Customize `configs/config.yaml` (optional)
- [ ] Set conservative limits for first run
- [ ] Run `go run main.go`
- [ ] Watch browser window (headless=false)
- [ ] Verify natural behavior
- [ ] Check database for results
- [ ] Review logs for errors
- [ ] Monitor LinkedIn account for warnings

---

## 📞 Need Help?

### Common Commands Reference

```powershell
# Install dependencies
go mod download

# Run application
go run main.go

# Build executable
go build -o linkedin-bot.exe main.go

# Run with debug logging
$env:LOG_LEVEL="debug"
go run main.go

# Check Go version
go version

# View Go environment
go env
```

### File Locations

- **Configuration**: `configs/config.yaml`
- **Credentials**: `.env`
- **Database**: `data/linkedin_bot.db`
- **Cookies**: `cookies.json`
- **Browser Data**: `browser-data/`
- **Logs**: Console output (stdout)

---

## ✅ Success Indicators

You'll know it's working when you see:

1. ✅ Browser opens automatically
2. ✅ LinkedIn login page loads
3. ✅ Credentials typed naturally (with pauses)
4. ✅ Login successful
5. ✅ Search results appear
6. ✅ Profiles visited with natural scrolling
7. ✅ Connection requests sent with notes
8. ✅ Database populated with records
9. ✅ No security challenges from LinkedIn
10. ✅ Daily stats displayed at end

**Happy Automating! 🚀**

Remember: Use responsibly and ethically. This tool is for educational purposes.
