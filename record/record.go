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

func wrapHeader(color int, prefix string) string {
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

func newBuilder(value ...any) *strings.Builder {
	var builder strings.Builder
	for index, arg := range value {
		for _, char := range fmt.Sprint(arg) {
			builder.WriteRune(char)
		}
		if index != len(value)-1 {
			builder.WriteRune(' ')
		}
	}
	return &builder
}

func Build(value ...any) string {
	return newBuilder(value...).String()
}

func Buildln(value ...any) string {
	p := newBuilder(value...)
	p.WriteRune('\n')
	return p.String()
}

func formatString(str string) string {
	if !Terminal {
		return str
	}
	var builder strings.Builder
	for index, char := range str {
		switch true {
		case char == '\n':
			builder.WriteRune('\r')
			builder.WriteRune('\n')

		case char == '\r':
			if index < len(str)-1 && str[index+1] == '\n' {
				continue
			}
			builder.WriteRune('\r')

		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func wrapLog(color int, prefix string, value ...any) string {
	return fmt.Sprintf("[%v]: %v", wrapHeader(color, prefix), Build(value...))
}

func wrapLogln(color int, prefix string, value ...any) string {
	return fmt.Sprintf("[%v]: %v", wrapHeader(color, prefix), Buildln(value...))
}

func Fprint(inout *os.File, value ...any) {
	fmt.Fprint(inout, formatString(Build(value...)))
}

func Print(value ...any) {
	fmt.Fprint(os.Stdout, formatString(Build(value...)))
}

func Println(value ...any) {
	fmt.Fprint(os.Stdout, formatString(Buildln(value...)))
}

func Printf(format string, value ...any) {
	fmt.Fprint(os.Stdout, formatString(fmt.Sprintf(format, value...)))
}

func Info(value ...any) {
	Fprint(os.Stdout, wrapLogln(Green, "INFO", value...))
}

func Warn(value ...any) {
	Fprint(os.Stdout, wrapLogln(Yellow, "WARN", value...))
}

func Debug(value ...any) {
	Fprint(os.Stdout, wrapLogln(Blue, "DEBUG", value...))
}

func Error(value ...any) {
	Fprint(os.Stderr, wrapLogln(Red, "ERROR", value...))
	os.Exit(1)
}

func ErrorNoExit(value ...any) {
	Fprint(os.Stderr, wrapLogln(Red, "ERROR", value...))
}

func RunningInfo(run func() error, value ...any) error {
	Fprint(os.Stdout, wrapLog(Green, "INFO", value...))
	err := run()
	Fprint(os.Stdout, "\n")
	if err != nil {
		return err
	}
	return nil
}

func FmtError(v ...any) error {
	return errors.New(Build(v...))
}
