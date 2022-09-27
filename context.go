package jrr

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

const abortIndex int8 = math.MaxInt8 >> 1

type Context struct {
	// origin object
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int

	// middlewares 本次context上下文所拥有的中间件
	handlers []HandlerFunc
	index    int8 // 记录当前context执行到第几个中间件

	// 框架引擎指针
	engine *Engine
}

func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) Run() {
	c.index = -1
	c.Next()
}

// Next 执行上下文中handlers中index下一个handler
func (c *Context) Next() {
	c.index++
	for ; c.index < int8(len(c.handlers)); c.index++ {
		c.handlers[c.index](c)
	}
}

// Abort 中止执行返回
func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "application/plain")
	c.Status(code)
	_, _ = c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	newEncoder := json.NewEncoder(c.Writer)
	if err := newEncoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, _ = c.Writer.Write(data)
}

func (c *Context) Fail(code int, data string) {
	c.Status(code)
	_, _ = c.Writer.Write([]byte(data))
}

// HTML 渲染html模版
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}
