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

func wrapColor(color int, value ...any) string {
	return fmt.Sprintf("\033[1;%vm", color) + fmt.Sprint(value...) + "\033[1;0m"
}

func wrapLog(value ...any) string {
	return "[" + fmt.Sprint(value...) + "]: "
}

func Info(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Green, "INFO")}
	fmt.Fprint(os.Stdout, wrapLog(strings.Join(arr, " ")))
	fmt.Println(value...)
}

func Warn(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Yellow, "WARN")}
	fmt.Fprint(os.Stdout, wrapLog(strings.Join(arr, " ")))
	fmt.Println(value...)
}


func Error(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Red, "ERROR")}
	fmt.Fprint(os.Stderr, wrapLog(strings.Join(arr, " ")))
	fmt.Fprintln(os.Stderr, value...)
	os.Exit(1)
}

func Debug(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Blue, "DEBUG")}
	fmt.Fprint(os.Stdout, wrapLog(strings.Join(arr, " ")))
	fmt.Fprintln(os.Stdout, value...)
}

func ErrorNoExit(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Red, "ERROR")}
	fmt.Fprint(os.Stderr, wrapLog(strings.Join(arr, " ")))
	fmt.Fprintln(os.Stderr, value...)
}

func InfoNoWrap(value ...any) {
	arr := []string{wrapColor(Blue, time.Now().Format(time.DateTime)), wrapColor(Green, "INFO")}
	fmt.Fprint(os.Stdout, wrapLog(strings.Join(arr, " ")))
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
