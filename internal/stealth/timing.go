package stealth

import (
	"math/rand"
	"time"
)

// TimingController handles randomized timing patterns
type TimingController struct {
	actionDelayMin  int
	actionDelayMax  int
	thinkTimeMin    int
	thinkTimeMax    int
	readingSpeedWPM int
	rand            *rand.Rand
}

// NewTimingController creates a new timing controller
func NewTimingController(actionDelayMin, actionDelayMax, thinkTimeMin, thinkTimeMax, readingSpeedWPM int) *TimingController {
	return &TimingController{
		actionDelayMin:  actionDelayMin,
		actionDelayMax:  actionDelayMax,
		thinkTimeMin:    thinkTimeMin,
		thinkTimeMax:    thinkTimeMax,
		readingSpeedWPM: readingSpeedWPM,
		rand:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ActionDelay returns a random delay between actions
func (t *TimingController) ActionDelay() time.Duration {
	delay := t.actionDelayMin + t.rand.Intn(t.actionDelayMax-t.actionDelayMin+1)
	return time.Duration(delay) * time.Second
}

// ThinkTime returns a random "think time" before an action
func (t *TimingController) ThinkTime() time.Duration {
	delay := t.thinkTimeMin + t.rand.Intn(t.thinkTimeMax-t.thinkTimeMin+1)
	return time.Duration(delay) * time.Second
}

// ReadingTime calculates reading time based on word count
func (t *TimingController) ReadingTime(wordCount int) time.Duration {
	if wordCount == 0 {
		return 0
	}

	// Calculate base reading time
	minutes := float64(wordCount) / float64(t.readingSpeedWPM)
	seconds := minutes * 60

	// Add some variation (±20%)
	variation := 0.2
	factor := 1 + (t.rand.Float64()*2-1)*variation

	return time.Duration(seconds*factor) * time.Second
}

// ShortPause returns a short random pause
func (t *TimingController) ShortPause() time.Duration {
	delay := 300 + t.rand.Intn(700)
	return time.Duration(delay) * time.Millisecond
}

// MediumPause returns a medium random pause
func (t *TimingController) MediumPause() time.Duration {
	delay := 1000 + t.rand.Intn(2000)
	return time.Duration(delay) * time.Millisecond
}

// LongPause returns a long random pause
func (t *TimingController) LongPause() time.Duration {
	delay := 3000 + t.rand.Intn(5000)
	return time.Duration(delay) * time.Millisecond
}

// RandomPause returns a random pause of varying length
func (t *TimingController) RandomPause() time.Duration {
	pauseType := t.rand.Intn(3)
	switch pauseType {
	case 0:
		return t.ShortPause()
	case 1:
		return t.MediumPause()
	default:
		return t.LongPause()
	}
}

// Wait waits for the specified duration
func (t *TimingController) Wait(duration time.Duration) {
	time.Sleep(duration)
}

// WaitActionDelay waits for a random action delay
func (t *TimingController) WaitActionDelay() {
	time.Sleep(t.ActionDelay())
}

// WaitThinkTime waits for a random think time
func (t *TimingController) WaitThinkTime() {
	time.Sleep(t.ThinkTime())
}
