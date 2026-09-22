package record

import (
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
	EnableColor bool = false
)

func wrapColor(color int, value ...any) string {
	if !EnableColor { return fmt.Sprint(value...) }
	return fmt.Sprintf("\033[1;%vm", color) + fmt.Sprint(value...) + "\033[1;0m"
}

func wrapPrefix(value ...any) string {
	return "[" + fmt.Sprint(value...) + "]: "
}

func Info(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Green, "INFO")}
	fmt.Fprint(os.Stdout, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Println(value...)
}

func Warn(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Yellow, "WARN")}
	fmt.Fprint(os.Stdout, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Println(value...)
}

func Error(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Red, "ERROR")}
	fmt.Fprint(os.Stderr, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Fprintln(os.Stderr, value...)
	os.Exit(1)
}

func Debug(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Blue, "DEBUG")}
	fmt.Fprint(os.Stdout, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Fprintln(os.Stdout, value...)
	time.Sleep(3*time.Second)
}

func ErrorNoExit(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Red, "ERROR")}
	fmt.Fprint(os.Stderr, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Fprintln(os.Stderr, value...)
}

func InfoNoWrap(value ...any) {
	prefix := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Green, "INFO")}
	fmt.Fprint(os.Stdout, wrapPrefix(strings.Join(prefix, " ")))
	fmt.Fprint(os.Stdout, value...)
}

func RunningInfo(run func() error, value ...any) error {
	InfoNoWrap(value...)
	defer fmt.Fprintf(os.Stdout, "\n")
	if err := run(); err != nil {
		return err
	}
	return nil
}
