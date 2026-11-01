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
	Terminal        bool = false
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

func joinValue(value ...any) string {
	var content string
	for index, arg := range value {
		if index != 0 {
			content += " "
		}
		content += fmt.Sprintf("%v", arg)
	}
	return content
}

func lineEnding() string {
	if Terminal {
		return "\r\n"
	}
	return "\n"
}

// normalizeEOL rewrites every line break in text to the one the current
// output mode expects, so multi line output stays aligned with column 0.
func normalizeEOL(text string) string {
	if !Terminal {
		return text
	}

	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		switch text[index] {
		case '\r':
			if index+1 < len(text) && text[index+1] == '\n' {
				continue
			}
			builder.WriteByte('\r')
		case '\n':
			builder.WriteString("\r\n")
		default:
			builder.WriteByte(text[index])
		}
	}
	return builder.String()
}

func wrapLog(color int, prefix string, value ...any) string {
	return fmt.Sprintf("[%v]: %v", wrapHander(color, prefix), joinValue(value...))
}

func wrapLogln(color int, prefix string, value ...any) string {
	return fmt.Sprintf("[%v]: %v%v", wrapHander(color, prefix), joinValue(value...), lineEnding())
}

func Print(value ...any) {
	fmt.Print(normalizeEOL(joinValue(value...)))
}

func Println(value ...any) {
	fmt.Print(normalizeEOL(joinValue(value...)) + lineEnding())
}

func Printf(format string, value ...any) {
	Print(fmt.Sprintf(format, value...))
}

func Info(value ...any) {
	Print(wrapLogln(Green, "INFO", value...))
}

func Warn(value ...any) {
	Print(wrapLogln(Yellow, "WARN", value...))
}

func Error(value ...any) {
	Print(wrapLogln(Red, "ERROR", value...))
	os.Exit(1)
}

func Debug(value ...any) {
	Print(wrapLogln(Blue, "DEBUG", value...))
}

func ErrorNoExit(value ...any) {
	Print(wrapLogln(Red, "ERROR", value...))
}

func RunningInfo(run func() error, value ...any) error {
	Print(wrapLog(Green, "INFO", value...))
	err := run()
	Println()
	if err != nil {
		return err
	}
	return nil
}

func FmtError(v ...any) error {
	return errors.New(joinValue(v...))
}
