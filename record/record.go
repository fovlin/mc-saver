package record

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Red = 31 + iota
	Green
	Yellow
	Blue
)

var (
	EnableColor     bool = false
	EnableTimestamp bool = false
)

func wrapColor(color int, value ...any) string {
	if !EnableColor {
		return fmt.Sprint(value...)
	}
	return fmt.Sprintf("\033[1;%vm", color) + fmt.Sprint(value...) + "\033[1;0m"
}

func wrapHander(color int, prefix string) string {
	var prefixs []string
	switch true {
	case !EnableTimestamp:
		prefixs = []string{
			wrapColor(color, prefix),
		}
	default:
		prefixs = []string{
			wrapColor(Blue, time.Now().Format(time.DateTime)),
			wrapColor(color, prefix),
		}
	}
	return strings.Join(prefixs, " ")
}

func wrapLog(color int, prefix string, value ...any) string {
	var content string
	for index, arg := range value {
		if index != 0 {
			content += " "
		}
		content += fmt.Sprintf("%v", arg)
	}
	return fmt.Sprintf("[%v]: %v", wrapHander(color, prefix), content)
}

func wrapLogln(color int, prefix string, value ...any) string {
	return fmt.Sprintf("[%v]: %v", wrapHander(color, prefix), fmt.Sprintln(value...))
}

func Info(value ...any) {
	fmt.Print(wrapLogln(Green, "INFO", value...))
}

func Warn(value ...any) {
	fmt.Print(wrapLogln(Yellow, "WARN", value...))
}

func Error(value ...any) {
	fmt.Print(wrapLogln(Red, "ERROR", value...))
	os.Exit(1)
}

func Debug(value ...any) {
	fmt.Print(wrapLogln(Blue, "DEBUG", value...))
}

func ErrorNoExit(value ...any) {
	fmt.Print(wrapLogln(Red, "ERROR", value...))
}

func RunningInfo(run func() error, value ...any) error {
	fmt.Printf("%s", wrapLog(Green, "INFO", value...))
	err := run()
	fmt.Print("\n")
	if err != nil {
		return err
	}
	return nil
}

func FmtError(v ...any) error {
	fmtString, _ := strings.CutSuffix(fmt.Sprintln(v...), "\n")
	return errors.New(fmtString)
}
