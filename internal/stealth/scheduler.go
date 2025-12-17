package stealth

import (
	"math/rand"
	"time"
)

// Scheduler handles activity scheduling
type Scheduler struct {
	businessHoursStart int
	businessHoursEnd   int
	timezone           *time.Location
	weekendActivity    bool
	breakDurationMin   int
	breakDurationMax   int
	breakProbability   float64
	rand               *rand.Rand
}

// NewScheduler creates a new scheduler
func NewScheduler(businessHoursStart, businessHoursEnd int, timezone string, weekendActivity bool, breakDurationMin, breakDurationMax int, breakProbability float64) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		businessHoursStart: businessHoursStart,
		businessHoursEnd:   businessHoursEnd,
		timezone:           loc,
		weekendActivity:    weekendActivity,
		breakDurationMin:   breakDurationMin,
		breakDurationMax:   breakDurationMax,
		breakProbability:   breakProbability,
		rand:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// IsBusinessHours checks if current time is within business hours
func (s *Scheduler) IsBusinessHours() bool {
	now := time.Now().In(s.timezone)
	hour := now.Hour()

	// Check if weekend
	if !s.weekendActivity && (now.Weekday() == time.Saturday || now.Weekday() == time.Sunday) {
		return false
	}

	// Check if within business hours
	return hour >= s.businessHoursStart && hour < s.businessHoursEnd
}

// WaitForBusinessHours waits until business hours
func (s *Scheduler) WaitForBusinessHours() {
	for !s.IsBusinessHours() {
		now := time.Now().In(s.timezone)
		
		// Calculate next business hour
		var nextBusinessTime time.Time
		
		// If weekend, wait until Monday
		if now.Weekday() == time.Saturday {
			nextBusinessTime = time.Date(now.Year(), now.Month(), now.Day()+2, s.businessHoursStart, 0, 0, 0, s.timezone)
		} else if now.Weekday() == time.Sunday {
			nextBusinessTime = time.Date(now.Year(), now.Month(), now.Day()+1, s.businessHoursStart, 0, 0, 0, s.timezone)
		} else {
			// Weekday - wait until business hours start
			if now.Hour() < s.businessHoursStart {
				nextBusinessTime = time.Date(now.Year(), now.Month(), now.Day(), s.businessHoursStart, 0, 0, 0, s.timezone)
			} else {
				// After business hours - wait until next day
				nextBusinessTime = time.Date(now.Year(), now.Month(), now.Day()+1, s.businessHoursStart, 0, 0, 0, s.timezone)
			}
		}

		waitDuration := time.Until(nextBusinessTime)
		time.Sleep(waitDuration)
	}
}

// ShouldTakeBreak determines if a break should be taken
func (s *Scheduler) ShouldTakeBreak() bool {
	return s.rand.Float64() < s.breakProbability
}

// TakeBreak takes a random break
func (s *Scheduler) TakeBreak() {
	duration := s.breakDurationMin + s.rand.Intn(s.breakDurationMax-s.breakDurationMin+1)
	time.Sleep(time.Duration(duration) * time.Minute)
}

// GetRandomStartTime returns a random time within business hours for starting activity
func (s *Scheduler) GetRandomStartTime() time.Time {
	now := time.Now().In(s.timezone)
	
	// Random hour within business hours
	hour := s.businessHoursStart + s.rand.Intn(s.businessHoursEnd-s.businessHoursStart)
	minute := s.rand.Intn(60)

	startTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.timezone)

	// If the time has passed today, schedule for tomorrow
	if startTime.Before(now) {
		startTime = startTime.Add(24 * time.Hour)
	}

	return startTime
}

// WaitUntil waits until a specific time
func (s *Scheduler) WaitUntil(targetTime time.Time) {
	duration := time.Until(targetTime)
	if duration > 0 {
		time.Sleep(duration)
	}
}
