package main

import (
	"fmt"

	"acovia.net/option"
)

func main() {
	var op string
	option.LongStringVar(&op, "test-var", "default")
	option.Parse()
	fmt.Println(option.Args)
	fmt.Println(op)
}