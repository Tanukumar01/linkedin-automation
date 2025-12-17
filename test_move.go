package main

import (
	"github.com/go-rod/rod"
)

func main() {
	p := rod.New().MustConnect().MustPage("Change Me")
	p.Mouse.Move(100, 100)
}
