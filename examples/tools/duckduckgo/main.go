package main

import (
	"fmt"

	"github.com/maksymenkoml/lingoose/tool/duckduckgo"
)

func main() {

	t := duckduckgo.New().WithMaxResults(5)
	f := t.Fn().(duckduckgo.FnPrototype)

	fmt.Println(f(duckduckgo.Input{Query: "Simone Vellei"}))
}
