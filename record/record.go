// Package record 提供带颜色的控制台日志输出（INFO/WARN/ERROR）。
package record

import (
	"fmt"
	"os"
	"time"
)


func Info(value ...any) {
	fmt.Fprint(os.Stdout, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;32mINFO\033[0m]: ")
	fmt.Println(value...)
}

func Warn(value ...any) {
	fmt.Fprint(os.Stdout, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;33mWARN\033[0m]: ")
	fmt.Println(value...)
}

func Error(value ...any) {
	fmt.Fprint(os.Stderr, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;31mERROR\033[0m]: ")
	fmt.Println(value...)
}

func Debug(value ...any) {
	fmt.Fprint(os.Stdout, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;34mDebug\033[0m]: ")
	fmt.Println(value...)
	os.Exit(1)
}

func ErrorNoExit(value ...any) {
	fmt.Fprint(os.Stderr, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;31mERROR\033[0m]: ")
	fmt.Println(value...)
}

func InfoNoWrap(format string, value ...any) {
	fmt.Fprint(os.Stdout, "[\033[1;34m" + time.Now().Format(time.DateTime) + " \033[1;32mINFO\033[0m]: ")
	fmt.Fprint(os.Stdout, value...)
}

func Wrap() {
	fmt.Fprintf(os.Stdout, "\n")
}

func RunningInfo(run func() error, format string, value ...any) error {
	InfoNoWrap(format, value...)
	defer Wrap()
	if err := run(); err != nil {
		return err
	}
	return nil
}
