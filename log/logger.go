package log

import (
	"fmt"
	"github.com/fatih/color"
)

var info = color.New(color.FgBlue).Sprint("INFO")
var error = color.New(color.FgRed).Sprint("ERROR")

func INFO(a ...any) {
	str := fmt.Sprint(a...)
	fmt.Println(info, str)
}

func ERROR(a ...any) {
	str := fmt.Sprint(a...)
	fmt.Println(error, str)
}

func INFOf(format string, a ...any) {
	str := fmt.Sprintf(format, a...)
	fmt.Println(info, str)
}

func ERRORf(format string, a ...any) {
	str := fmt.Sprintf(format, a...)
	fmt.Println(error, str)
}
