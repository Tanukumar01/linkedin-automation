package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// Typer handles realistic typing simulation
type Typer struct {
	wpmMin           int
	wpmMax           int
	typoProbability  float64
	pauseProbability float64
	rand             *rand.Rand
}

// NewTyper creates a new typer
func NewTyper(wpmMin, wpmMax int, typoProbability, pauseProbability float64) *Typer {
	return &Typer{
		wpmMin:           wpmMin,
		wpmMax:           wpmMax,
		typoProbability:  typoProbability,
		pauseProbability: pauseProbability,
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// TypeText types text with human-like behavior
func (t *Typer) TypeText(page *rod.Page, element *rod.Element, text string) error {
	// Focus on the element
	if err := element.Focus(); err != nil {
		return err
	}

	// Calculate typing speed (characters per minute to milliseconds per character)
	wpm := t.wpmMin + t.rand.Intn(t.wpmMax-t.wpmMin+1)
	cpm := wpm * 5 // Average word length is 5 characters
	msPerChar := 60000 / cpm

	for i, char := range text {
		// Random pause before some characters
		if t.rand.Float64() < t.pauseProbability {
			pauseDuration := time.Duration(200+t.rand.Intn(500)) * time.Millisecond
			time.Sleep(pauseDuration)
		}

		// Simulate typo
		if t.rand.Float64() < t.typoProbability && i > 0 {
			// Type a wrong character
			wrongChar := t.getRandomChar()
			page.Keyboard.Type(input.Key(wrongChar))
			time.Sleep(time.Duration(msPerChar+t.rand.Intn(100)) * time.Millisecond)

			// Backspace to correct
			page.Keyboard.Press(input.Backspace)
			time.Sleep(time.Duration(msPerChar+t.rand.Intn(100)) * time.Millisecond)
		}

		// Type the correct character
		page.Keyboard.Type(input.Key(char))

		// Variable delay between characters
		delay := msPerChar + t.rand.Intn(msPerChar/2) - msPerChar/4
		time.Sleep(time.Duration(delay) * time.Millisecond)

		// Longer pause after punctuation
		if char == '.' || char == ',' || char == '!' || char == '?' {
			time.Sleep(time.Duration(100+t.rand.Intn(300)) * time.Millisecond)
		}

		// Pause between words
		if char == ' ' {
			time.Sleep(time.Duration(50+t.rand.Intn(150)) * time.Millisecond)
		}
	}

	return nil
}

// getRandomChar returns a random character for typo simulation
func (t *Typer) getRandomChar() rune {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	return chars[t.rand.Intn(len(chars))]
}

// ClearAndType clears an input field and types new text
func (t *Typer) ClearAndType(page *rod.Page, element *rod.Element, text string) error {
	// Focus on element
	if err := element.Focus(); err != nil {
		return err
	}

	// Select all and delete
	page.Keyboard.Press(input.ControlLeft)
	page.Keyboard.Type(input.Key('a'))
	page.Keyboard.Release(input.ControlLeft)
	time.Sleep(time.Duration(50+t.rand.Intn(100)) * time.Millisecond)

	page.Keyboard.Press(input.Backspace)
	time.Sleep(time.Duration(100+t.rand.Intn(200)) * time.Millisecond)

	// Type new text
	return t.TypeText(page, element, text)
}
