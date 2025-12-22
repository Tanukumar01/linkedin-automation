package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	userDataDir := filepath.Join(os.TempDir(), "linkedin-bot-browser-data-test")
	os.MkdirAll(userDataDir, 0755)

	l := launcher.New().
		Headless(true).
		UserDataDir(userDataDir).
		Leakless(false).
		NoSandbox(true).
		Set("disable-gpu").
		Logger(os.Stdout)

	if path, exists := launcher.LookPath(); exists {
		l.Bin(path)
	}

	fmt.Printf("Starting with user data dir: %s\n", userDataDir)
	url, err := l.Launch()
	if err != nil {
		fmt.Printf("Launch failed: %v\n", err)
		return
	}
	fmt.Printf("Debug URL: %s\n", url)
}
