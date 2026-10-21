package option

import (
	"os"
	"strings"
)

type VarType uint8
// 1 - Bool
// 2 - Int
// 3 - String

var (
	StringOptions = make(map[string]StringValue)
	BoolOptions = make(map[string]BoolValue)
	Args []string
)

type StringValue struct {
	Pointer *string
	Value string
}

type BoolValue struct {
	Pointer *bool
	Value bool
}

func StringVar(v *string, shortName string, longName string, defaultValue string) {
	ShortStringVar(v, shortName, defaultValue)
	LongStringVar(v, longName, defaultValue)
}

func ShortStringVar(v *string, name string, defaultValue string) {
	StringOptions["-"+name] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func LongStringVar(v *string, name string, defaultValue string) {
	StringOptions["--"+name] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func BoolVar(v *bool, shortName string, longName string) {
	ShortBoolVar(v, shortName)
	LongBoolVar(v, longName)
}

func ShortBoolVar(v *bool, name string) {
	BoolOptions["-"+name] = BoolValue{
		Pointer: v,
		Value: false,
	}
}

func LongBoolVar(v *bool, name string) {
	BoolOptions["--"+name] = BoolValue{
		Pointer: v,
		Value: false,
	}
}

func Parse() {
	scan()
	for _, v := range StringOptions {
		if v.Pointer != nil {
			*v.Pointer = v.Value
		}
	}
	for _, v := range BoolOptions {
		if v.Pointer != nil {
			*v.Pointer = v.Value
		}
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
			storeStringVar(targetName, arg)
			isOption = false

		case strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"):
			switch checkVarType(arg) {
			case 3:
				targetName = arg
				isOption = true

			case 1:
				storeBoolVar(arg)
			}

		default:
			Args = append(Args, arg)
		}
	}
}

func checkVarType(name string) VarType {
	var vType VarType
	_, stringOk := StringOptions[name]
	_, boolOk := BoolOptions[name]

	switch true {
	case stringOk:
		vType = 3

	case boolOk:
		vType = 1
	}

	return vType
}

func storeStringVar(name string, arg string) {
	stringValue, ok := StringOptions[name]
	if ok {
		stringValue.Value = arg
		StringOptions[name] = stringValue
	}
}

func storeBoolVar(name string) {
	boolValue, ok := BoolOptions[name]
	if ok {
		boolValue.Value = true
		BoolOptions[name] = boolValue
	}
}