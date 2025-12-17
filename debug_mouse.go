package main

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/go-rod/rod"
)

func main() {
	types := []interface{}{
		&rod.Mouse{},
		&rod.Page{},
		&rod.Element{},
	}

	for _, t := range types {
		typ := reflect.TypeOf(t)
		fmt.Printf("\n--- Methods for %s ---\n", typ)
		var methods []string
		for i := 0; i < typ.NumMethod(); i++ {
			methods = append(methods, typ.Method(i).Name)
		}
		sort.Strings(methods)
		for _, m := range methods {
			fmt.Println(m)
		}
	}
}
