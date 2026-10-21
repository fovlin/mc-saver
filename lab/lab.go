package main

import (
	"fmt"

	"acovia.net/option"
)

func main() {
	var op bool
	option.LongBoolVar(&op, "test-var")
	option.Parse()
	fmt.Println(option.Args)
	fmt.Println(op)
}