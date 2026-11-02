package opt

import (
	"fmt"
	"os"
	"strings"
)

var (
	null     string = ""
	Options         = make(map[string]Value)
	args     []string
	funcList []func()
)

type Value interface {
	NextIsOption() bool
	SetValue(string)
}

type StringFlag struct {
	NextOption bool
	Value      *string
}

type BoolFlag struct {
	NextOption bool
	Value      *bool
}

func (stringValue StringFlag) SetValue(arg string) {
	*stringValue.Value = arg
}

func (boolValue BoolFlag) SetValue(arg string) {
	*boolValue.Value = true
}

func (stringValue StringFlag) NextIsOption() bool {
	return stringValue.NextOption
}

func (boolValue BoolFlag) NextIsOption() bool {
	return boolValue.NextOption
}

func StringVar(v *string, shortName string, longName string, defaultValue string) {
	if len(shortName) != 0 {
		ShortStringVar(v, shortName, defaultValue)
	}
	if len(longName) != 0 {
		LongStringVar(v, longName, defaultValue)
	}
}

func ShortStringVar(v *string, name string, defaultValue string) {
	if len(name) != 1 {
		panic("the length of the short parameter name must be equal to 1")
	}
	name = "-" + name
	_, ok := Options[name]
	if ok {
		panic("duplicate registration parameter: " + name)
	}
	*v = defaultValue
	Options[name] = StringFlag{
		NextOption: true,
		Value:      v,
	}
}

func LongStringVar(v *string, name string, defaultValue string) {
	if len(name) == 0 {
		panic("the length of the long parameter name must be greater than 0")
	}
	name = "--" + name
	_, ok := Options[name]
	if ok {
		panic("duplicate registration parameter: " + name)
	}
	*v = defaultValue
	Options[name] = StringFlag{
		NextOption: true,
		Value:      v,
	}
}

func BoolSwitch(v *bool, shortName string, longName string) {
	if len(shortName) != 0 {
		ShortBoolSwitch(v, shortName)
	}
	if len(longName) != 0 {
		LongBoolSwitch(v, longName)
	}
}

func ShortBoolSwitch(v *bool, name string) {
	if len(name) != 1 {
		panic("the length of the short parameter name must be equal to 1")
	}
	name = "-" + name
	_, ok := Options[name]
	if ok {
		panic("duplicate registration parameter: " + name)
	}
	Options[name] = BoolFlag{
		Value: v,
	}
}

func LongBoolSwitch(v *bool, name string) {
	if len(name) == 0 {
		panic("the length of the long parameter name must be greater than 0")
	}
	name = "--" + name
	_, ok := Options[name]
	if ok {
		panic("duplicate registration parameter: " + name)
	}
	Options[name] = BoolFlag{
		Value: v,
	}
}

func BoolFunc(shortName string, longName string, function func()) {
	var boolValue bool
	if shortName != "" {
		ShortBoolSwitch(&boolValue, shortName)
	}
	if longName != "" {
		LongBoolSwitch(&boolValue, longName)
	}

	funcList = append(funcList, func() {
		if boolValue {
			function()
		}
	})
}

func ShortBoolFunc(shortName string, function func()) {
	BoolFunc(shortName, "", function)
}

func LongBoolFunc(longName string, function func()) {
	BoolFunc("", longName, function)
}

func Parse() error {
	args = nil

	err := scan()
	if err != nil {
		return err
	}

	for _, function := range funcList {
		function()
	}

	funcList = nil
	Options = make(map[string]Value)
	return nil
}

func Args() []string {
	return args
}

func scan() error {
	var lastValue Value
	for _, arg := range os.Args[1:] {
		switch true {
		case lastValue != nil && lastValue.NextIsOption():
			lastValue.SetValue(arg)
			lastValue = nil

		case strings.HasPrefix(arg, "--") && strings.ContainsRune(arg, '='):
			fields := strings.SplitN(arg, "=", 2)
			value, ok := Options[fields[0]]
			switch true {
			case !ok:
				return fmt.Errorf("unknown option: %v", arg)

			case value.NextIsOption():
				value.SetValue(fields[1])

			default:
				value.SetValue(null)
			}

		case strings.HasPrefix(arg, "--"):
			value, ok := Options[arg]
			switch true {
			case !ok:
				return fmt.Errorf("unknown option: %v", arg)

			case value.NextIsOption():
				lastValue = value

			default:
				value.SetValue(null)
			}

		case strings.HasPrefix(arg, "-"):
			arg, _ = strings.CutPrefix(arg, "-")
			for n, subArg := range arg {
				value, ok := Options["-"+string(subArg)]
				switch true {
				case !ok:
					return fmt.Errorf("unknown option: %v", arg)

				case n == len(arg)-1 && value.NextIsOption():
					lastValue = value

				default:
					value.SetValue(null)
				}
			}

		default:
			args = append(args, arg)

		}
	}
	return nil
}

func ParseLine(cmd string) []string {
	var str string
	var result []string
	var escape, sinQuo, douQuo bool

	for _, char := range cmd {
		switch true {
		case escape:
			escape = false
			str += string(char)

		case char == '\\' && !sinQuo:
			escape = true

		case char == '\'' && !douQuo:
			sinQuo = !sinQuo

		case char == '"' && !sinQuo:
			douQuo = !douQuo

		case char == ' ' || char == '\t' && !sinQuo && !douQuo:
			if len(str) != 0 {
				result = append(result, str)
			}
			str = ""

		default:
			str += string(char)
		}
	}

	if len(str) != 0 {
		result = append(result, str)
	}

	return result
}
