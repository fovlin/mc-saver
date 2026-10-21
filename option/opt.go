package option

import (
	"os"
	"strings"
)

var (
	Options = make(map[string]Value)
	Args []string
)

type Value interface {
	SetValue()
}

type StringValue struct {
	Pointer *string
	Value string
}

type BoolValue struct {
	Pointer *bool
	Value bool
}

func (stringValue StringValue) SetValue() {
	*stringValue.Pointer = stringValue.Value
}

func (boolValue BoolValue) SetValue() {
	*boolValue.Pointer = boolValue.Value
}

func StringVar(v *string, shortName string, longName string, defaultValue string) {
	ShortStringVar(v, shortName, defaultValue)
	LongStringVar(v, longName, defaultValue)
}

func ShortStringVar(v *string, name string, defaultValue string) {
	Options["-"+name] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func LongStringVar(v *string, name string, defaultValue string) {
	Options["--"+name] = StringValue{
		Pointer: v,
		Value: defaultValue,
	}
}

func BoolVar(v *bool, shortName string, longName string) {
	ShortBoolVar(v, shortName)
	LongBoolVar(v, longName)
}

func LongBoolVar(v *bool, name string) {
	Options["--"+name] = BoolValue{
		Pointer: v,
	}
}

func ShortBoolVar(v *bool, name string) {
	Options["-"+name] = BoolValue{
		Pointer: v,
	}
}

func Parse() {
	scan()
	for _, v := range Options {
		v.SetValue()
	}
}

func parseValue(arg string) (bool, string) {
	nextName := ""
	nextOption := true
	value, ok := Options[arg]
	if !ok {
		return true, nextName
	}

	switch v := value.(type) {
	case StringValue:
		nextOption = true
		nextName = arg

	case BoolValue:
		v.Value = true
		Options[arg] = v
		nextOption = false

	}

	return nextOption, nextName
}

func reParseValue(arg string, nextIsOption bool, nextName string) (bool, string) {
	value, _ := Options[nextName]

	switch v := value.(type) {
	case StringValue:
		v.Value = arg
		Options[nextName] = v
	}

	nextName = ""
	nextIsOption = false

	return nextIsOption, nextName
}

func scan() {
	var nextName string
	var nextIsOption bool
	if len(os.Args) < 2 {
		return
	}

	for _, arg := range os.Args[1:] {
		switch true {
		case nextIsOption:
			nextIsOption, nextName = reParseValue(arg, nextIsOption, nextName)

		case strings.HasPrefix(arg, "-"):
			nextIsOption, nextName = parseValue(arg)

		default:
			Args = append(Args, arg)
		}
	}
}