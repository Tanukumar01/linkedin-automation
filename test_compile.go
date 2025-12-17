package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	p := rod.New().MustConnect().MustPage("Change Me")
	pt := proto.NewPoint(100, 100)
	err := p.Mouse.MoveAlong(pt)
	if err != nil {
		fmt.Println(err)
	}
}
