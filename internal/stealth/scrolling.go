package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// Scroller handles natural scrolling behavior
type Scroller struct {
	speedMin              int
	speedMax              int
	scrollBackProbability float64
	pauseProbability      float64
	rand                  *rand.Rand
}

// NewScroller creates a new scroller
func NewScroller(speedMin, speedMax int, scrollBackProb, pauseProb float64) *Scroller {
	return &Scroller{
		speedMin:              speedMin,
		speedMax:              speedMax,
		scrollBackProbability: scrollBackProb,
		pauseProbability:      pauseProb,
		rand:                  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ScrollDown scrolls down the page naturally
func (s *Scroller) ScrollDown(page *rod.Page, distance int) error {
	// Break scrolling into smaller chunks
	chunks := 5 + s.rand.Intn(10)
	chunkSize := distance / chunks

	for i := 0; i < chunks; i++ {
		// Calculate scroll amount with variation
		scrollAmount := chunkSize + s.rand.Intn(chunkSize/2) - chunkSize/4

		// Scroll
		err := page.Mouse.Scroll(0, float64(scrollAmount), chunks)
		if err != nil {
			return err
		}

		// Variable delay between scrolls
		speed := s.speedMin + s.rand.Intn(s.speedMax-s.speedMin+1)
		time.Sleep(time.Duration(speed) * time.Millisecond)

		// Random pause
		if s.rand.Float64() < s.pauseProbability {
			pauseDuration := time.Duration(500+s.rand.Intn(1500)) * time.Millisecond
			time.Sleep(pauseDuration)
		}

		// Random scroll back
		if s.rand.Float64() < s.scrollBackProbability {
			scrollBack := s.rand.Intn(scrollAmount / 2)
			page.Mouse.Scroll(0, float64(-scrollBack), 1)
			time.Sleep(time.Duration(200+s.rand.Intn(300)) * time.Millisecond)
		}
	}

	return nil
}

// ScrollToElement scrolls to make an element visible
func (s *Scroller) ScrollToElement(page *rod.Page, element *rod.Element) error {
	// Get element position using JS since Box() is not available
	yVal := page.MustEval(`(el) => {
		const rect = el.getBoundingClientRect();
		return rect.top + window.pageYOffset;
	}`, element)

	// Get viewport height
	viewport := page.MustEval(`() => window.innerHeight`).Int()

	// Calculate scroll distance
	currentScroll := page.MustEval(`() => window.pageYOffset`).Int()
	targetScroll := yVal.Int() - viewport/2

	distance := targetScroll - currentScroll

	if distance > 0 {
		return s.ScrollDown(page, distance)
	} else if distance < 0 {
		return s.ScrollUp(page, -distance)
	}

	return nil
}

// ScrollUp scrolls up the page naturally
func (s *Scroller) ScrollUp(page *rod.Page, distance int) error {
	// Break scrolling into smaller chunks
	chunks := 5 + s.rand.Intn(10)
	chunkSize := distance / chunks

	for i := 0; i < chunks; i++ {
		// Calculate scroll amount with variation
		scrollAmount := chunkSize + s.rand.Intn(chunkSize/2) - chunkSize/4

		// Scroll up (negative value)
		err := page.Mouse.Scroll(0, float64(-scrollAmount), chunks)
		if err != nil {
			return err
		}

		// Variable delay between scrolls
		speed := s.speedMin + s.rand.Intn(s.speedMax-s.speedMin+1)
		time.Sleep(time.Duration(speed) * time.Millisecond)

		// Random pause
		if s.rand.Float64() < s.pauseProbability {
			pauseDuration := time.Duration(500+s.rand.Intn(1500)) * time.Millisecond
			time.Sleep(pauseDuration)
		}
	}

	return nil
}

// ScrollToBottom scrolls to the bottom of the page
func (s *Scroller) ScrollToBottom(page *rod.Page) error {
	// Get page height
	pageHeight := page.MustEval(`() => document.body.scrollHeight`).Int()
	currentScroll := page.MustEval(`() => window.pageYOffset`).Int()

	distance := pageHeight - currentScroll

	return s.ScrollDown(page, distance)
}

// ScrollToTop scrolls to the top of the page
func (s *Scroller) ScrollToTop(page *rod.Page) error {
	currentScroll := page.MustEval(`() => window.pageYOffset`).Int()
	return s.ScrollUp(page, currentScroll)
}

// RandomScroll performs random scrolling behavior
func (s *Scroller) RandomScroll(page *rod.Page) error {
	// Random scroll direction
	if s.rand.Float64() < 0.5 {
		distance := 200 + s.rand.Intn(500)
		return s.ScrollDown(page, distance)
	} else {
		distance := 100 + s.rand.Intn(300)
		return s.ScrollUp(page, distance)
	}
}
