package jrr

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// 错误恢复

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf(trace(message))
				c.Fail(http.StatusInternalServerError, "interval server error")
			}
		}()
		c.Next()
	}
}

// trace print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller 掉过前三个调用

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
