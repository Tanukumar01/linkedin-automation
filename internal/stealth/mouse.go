package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Point represents a 2D point
type Point struct {
	X, Y float64
}

// MouseMover handles human-like mouse movements
type MouseMover struct {
	page              *rod.Page
	bezierPoints      int
	speedVariation    float64
	overshootProb     float64
	microCorrectionProb float64
	rand              *rand.Rand
}

// NewMouseMover creates a new mouse mover
func NewMouseMover(page *rod.Page, bezierPoints int, speedVariation, overshootProb, microCorrectionProb float64) *MouseMover {
	return &MouseMover{
		page:              page,
		bezierPoints:      bezierPoints,
		speedVariation:    speedVariation,
		overshootProb:     overshootProb,
		microCorrectionProb: microCorrectionProb,
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// MoveToElement moves the mouse to an element with human-like behavior
func (m *MouseMover) MoveToElement(element *rod.Element) error {
	// Get element position and size
	box, err := element.Box()
	if err != nil {
		return err
	}

	// Calculate target position (random point within element)
	targetX := box.X + m.rand.Float64()*box.Width
	targetY := box.Y + m.rand.Float64()*box.Height

	// Get current mouse position
	currentPos, err := m.getCurrentPosition()
	if err != nil {
		// If we can't get current position, start from a random point
		currentPos = Point{
			X: m.rand.Float64() * 1000,
			Y: m.rand.Float64() * 600,
		}
	}

	// Move to target with Bézier curve
	return m.moveToPoint(currentPos, Point{X: targetX, Y: targetY})
}

// MoveToPoint moves the mouse to a specific point
func (m *MouseMover) moveToPoint(start, end Point) error {
	// Generate Bézier curve points
	path := m.generateBezierPath(start, end)

	// Add overshoot if probability hits
	if m.rand.Float64() < m.overshootProb {
		overshoot := m.generateOvershoot(end)
		path = append(path, overshoot...)
	}

	// Move along the path
	for i, point := range path {
		// Calculate delay with speed variation
		baseDelay := 5 + m.rand.Intn(10)
		variation := int(float64(baseDelay) * m.speedVariation * (m.rand.Float64()*2 - 1))
		delay := time.Duration(baseDelay+variation) * time.Millisecond

		// Move mouse
		err := m.page.Mouse.MoveAlong(proto.NewPoint(point.X, point.Y))
		if err != nil {
			return err
		}

		// Add micro-corrections randomly
		if i > 0 && i < len(path)-1 && m.rand.Float64() < m.microCorrectionProb {
			correction := Point{
				X: point.X + (m.rand.Float64()*4 - 2),
				Y: point.Y + (m.rand.Float64()*4 - 2),
			}
			m.page.Mouse.MoveAlong(proto.NewPoint(correction.X, correction.Y))
			time.Sleep(delay / 2)
		}

		time.Sleep(delay)
	}

	return nil
}

// generateBezierPath generates a Bézier curve path between two points
func (m *MouseMover) generateBezierPath(start, end Point) []Point {
	// Generate control points
	controlPoints := make([]Point, m.bezierPoints)
	controlPoints[0] = start
	controlPoints[m.bezierPoints-1] = end

	// Generate intermediate control points with randomness
	for i := 1; i < m.bezierPoints-1; i++ {
		t := float64(i) / float64(m.bezierPoints-1)
		
		// Linear interpolation with random offset
		x := start.X + (end.X-start.X)*t
		y := start.Y + (end.Y-start.Y)*t
		
		// Add random offset perpendicular to the line
		distance := math.Sqrt(math.Pow(end.X-start.X, 2) + math.Pow(end.Y-start.Y, 2))
		maxOffset := distance * 0.2
		offset := (m.rand.Float64()*2 - 1) * maxOffset
		
		angle := math.Atan2(end.Y-start.Y, end.X-start.X) + math.Pi/2
		x += offset * math.Cos(angle)
		y += offset * math.Sin(angle)
		
		controlPoints[i] = Point{X: x, Y: y}
	}

	// Generate points along the Bézier curve
	numPoints := 20 + m.rand.Intn(30)
	path := make([]Point, numPoints)

	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		path[i] = m.bezierPoint(controlPoints, t)
	}

	return path
}

// bezierPoint calculates a point on a Bézier curve
func (m *MouseMover) bezierPoint(controlPoints []Point, t float64) Point {
	n := len(controlPoints) - 1
	var x, y float64

	for i := 0; i <= n; i++ {
		coefficient := m.binomialCoefficient(n, i) * math.Pow(1-t, float64(n-i)) * math.Pow(t, float64(i))
		x += coefficient * controlPoints[i].X
		y += coefficient * controlPoints[i].Y
	}

	return Point{X: x, Y: y}
}

// binomialCoefficient calculates binomial coefficient
func (m *MouseMover) binomialCoefficient(n, k int) float64 {
	if k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	result := 1.0
	for i := 0; i < k; i++ {
		result *= float64(n - i)
		result /= float64(i + 1)
	}

	return result
}

// generateOvershoot generates overshoot points
func (m *MouseMover) generateOvershoot(target Point) []Point {
	overshootDistance := 10 + m.rand.Float64()*30
	angle := m.rand.Float64() * 2 * math.Pi

	overshoot := Point{
		X: target.X + overshootDistance*math.Cos(angle),
		Y: target.Y + overshootDistance*math.Sin(angle),
	}

	// Return path from overshoot back to target
	return []Point{overshoot, target}
}

// getCurrentPosition gets the current mouse position
func (m *MouseMover) getCurrentPosition() (Point, error) {
	// Rod doesn't provide a way to get current mouse position
	// We'll track it ourselves or estimate
	return Point{X: 0, Y: 0}, nil
}

// HoverElement hovers over an element
func (m *MouseMover) HoverElement(element *rod.Element) error {
	if err := m.MoveToElement(element); err != nil {
		return err
	}

	// Stay hovered for a random duration
	hoverDuration := time.Duration(500+m.rand.Intn(1500)) * time.Millisecond
	time.Sleep(hoverDuration)

	return nil
}

// ClickElement clicks an element with human-like behavior
func (m *MouseMover) ClickElement(element *rod.Element) error {
	// Move to element
	if err := m.MoveToElement(element); err != nil {
		return err
	}

	// Small pause before clicking
	time.Sleep(time.Duration(100+m.rand.Intn(300)) * time.Millisecond)

	// Click
	if err := m.page.Mouse.Click(proto.InputMouseButtonLeft); err != nil {
		return err
	}

	// Small pause after clicking
	time.Sleep(time.Duration(100+m.rand.Intn(200)) * time.Millisecond)

	return nil
}

// RandomIdleMovement performs random idle mouse movements
func (m *MouseMover) RandomIdleMovement() error {
	// Get viewport size
	viewport := m.page.MustEval(`() => ({ width: window.innerWidth, height: window.innerHeight })`).Map()
	width := viewport["width"].Num()
	height := viewport["height"].Num()

	// Generate random target within viewport
	target := Point{
		X: m.rand.Float64() * width,
		Y: m.rand.Float64() * height,
	}

	currentPos := Point{
		X: m.rand.Float64() * width,
		Y: m.rand.Float64() * height,
	}

	return m.moveToPoint(currentPos, target)
}
