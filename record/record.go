// Package record 提供带颜色的控制台日志输出（INFO/WARN/ERROR）。
package record

import (
	"fmt"
	"os"
)

// Info 向标准输出打印绿色 INFO 日志
func Info(format string, value ...any) {
	fmt.Fprintf(os.Stdout, "[\033[1;32mINFO\033[0m]: "+format+"\n", value...)
}

// Warn 向标准输出打印黄色 WARN 日志
func Warn(format string, value ...any) {
	fmt.Fprintf(os.Stdout, "[\033[1;33mWARN\033[0m]: "+format+"\n", value...)
}

// Debug 向标准输出打印蓝色 DEBUG 日志
func Debug(format string, value ...any) {
	fmt.Fprintf(os.Stdout, "[\033[1;34mDEBUG\033[0m]: "+format+"\n", value...)
}

// Error 向标准错误打印红色 ERROR 日志
func Error(format string, value ...any) {
	fmt.Fprintf(os.Stderr, "[\033[1;31mERROR\033[0m]: "+format+"\n", value...)
}
