# Quick Fix for CGO Error

## The Problem
The `mattn/go-sqlite3` package requires CGO (C compiler), which is not available on your system.

## Solution: Install GCC for Windows

### Option 1: Install TDM-GCC (Recommended - Easiest)

1. **Download TDM-GCC**:
   - Visit: https://jmeubank.github.io/tdm-gcc/
   - Download: `tdm64-gcc-10.3.0-2.exe` (or latest version)

2. **Install**:
   - Run the installer
   - Choose "Create" (new installation)
   - Select default options
   - Complete installation

3. **Verify**:
   ```powershell
   gcc --version
   ```

4. **Restart PowerShell** and try again:
   ```powershell
   go run main.go
   ```

### Option 2: Install MinGW-w64

1. **Download**:
   - Visit: https://www.mingw-w64.org/downloads/
   - Or use: https://github.com/niXman/mingw-builds-binaries/releases

2. **Install and add to PATH**

3. **Verify**:
   ```powershell
   gcc --version
   ```

### Option 3: Use Pure Go SQLite (No CGO Required)

If you don't want to install GCC, we can switch to a pure Go SQLite implementation:

1. **Update go.mod**:
   Replace `github.com/mattn/go-sqlite3` with `modernc.org/sqlite`

2. **Update imports** in `internal/storage/db.go`:
   Replace:
   ```go
   _ "github.com/mattn/go-sqlite3"
   ```
   With:
   ```go
   _ "modernc.org/sqlite"
   ```

3. **Run**:
   ```powershell
   go mod tidy
   go run main.go
   ```

## Quick Test After Installing GCC

```powershell
# Add Go to PATH (if needed)
$env:Path += ";C:\Program Files\Go\bin"

# Run the application
go run main.go
```

## What You Should See

Once GCC is installed and the app runs successfully:

1. ✅ "Starting LinkedIn Automation Bot"
2. ✅ "Database initialized"
3. ✅ "Browser initialized"
4. ✅ Browser window opens
5. ✅ Navigates to LinkedIn
6. ✅ Logs in automatically

## Current Status

- ✅ Go installed (v1.25.5)
- ✅ Dependencies downloaded
- ✅ .env file configured
- ❌ GCC/CGO not available (needed for SQLite)

**Next Step**: Install TDM-GCC (Option 1 above) - takes 2-3 minutes
