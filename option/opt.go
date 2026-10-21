package option

import (
	"fmt"
	"os"
	"strings"
)

var (
	StringOptions = make(map[string]StringValue)
	Args []string
)

type StringValue struct {
	Pointer *string
	Value string
}

func ShortStringVar(v *string, name string, defaultValue string) {
	StringOptions[fmt.Sprint("-", name)] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func LongStringVar(v *string, name string, defaultValue string) {
	StringOptions[fmt.Sprint("--", name)] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func Parse() {
	scan()
	for _, v := range StringOptions {
		*v.Pointer = v.Value
	}
}

func scan() {
	var targetName string
	var isOption bool
	if len(os.Args) < 2 {
		return
	}

	for _, arg := range os.Args[1:] {
		switch true {
		case isOption:
			storeVar(targetName, arg)
			isOption = false

		case strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"):
			if checkVar(arg) {
				targetName = arg
				isOption = true
			}

		default:
			Args = append(Args, arg)
		}
	}
}

func checkVar(name string) bool {
	_, stringOk := StringOptions[name]
	return stringOk
}

func storeVar(name string, arg string) {
	v, ok := StringOptions[name]
	if ok {
		v.Value = arg
		StringOptions[name] = v
	}
}