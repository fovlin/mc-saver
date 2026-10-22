package opt

import (
	"fmt"
	"os"
	"strings"
)

var (
	Options = make(map[string]Value)
	allArgs []string
	Args []string
	funcList []func()
)

type Value interface {
	NextIsOption() bool
	SetValue(string)
}

type StringFlag struct {
	NextOption bool
	Value *string
}

type BoolFlag struct {
	NextOption bool
	Value *bool
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
	*v = defaultValue
	Options["-"+name] = StringFlag{
		NextOption: true,
		Value: v,
	}
}

func LongStringVar(v *string, name string, defaultValue string) {
	*v = defaultValue
	Options["--"+name] = StringFlag{
		NextOption: true,
		Value: v,
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

func LongBoolSwitch(v *bool, name string) {
	Options["--"+name] = BoolFlag{
		Value: v,
	}
}

func ShortBoolSwitch(v *bool, name string) {
	Options["-"+name] = BoolFlag{
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

func Parse() error {
	err := initArgs()
	if err != nil {
		return err
	}
	scan()
	for _, function := range funcList {
		function()
	}
	return nil
}

func scan() {
	clear(Args)
	var nextValue Value
	if len(os.Args) < 2 {
		return
	}

	for _, arg := range allArgs {
		switch true {
		case nextValue != nil && nextValue.NextIsOption():
			nextValue.SetValue(arg)
			nextValue = nil

		case strings.HasPrefix(arg, "-"):
			var ok bool
			nextValue, ok = Options[arg]
			if ok && !nextValue.NextIsOption() {
				nextValue.SetValue("")
			}

		default:
			Args = append(Args, arg)
		}
	}
}

func initArgs() error {
	for _, arg := range os.Args[1:] {
		switch true {
		case strings.HasPrefix(arg, "-") && strings.ContainsRune(arg, '='):
			fields := strings.SplitN(arg, "=", 2)
			value, ok := Options[fields[0]]
			switch true {
			case !ok:
				return fmt.Errorf("unknown option: %v", arg)

			case value.NextIsOption():
				allArgs = append(allArgs, fields...)

			default:
				allArgs = append(allArgs, fields[0])
			}

		case strings.HasPrefix(arg, "-"):
			_, ok := Options[arg]
			if !ok {
				return fmt.Errorf("unknown option: %v", arg)
			}
			allArgs = append(allArgs, arg)

		default:
			allArgs = append(allArgs, arg)

		}

	}

	return nil
}