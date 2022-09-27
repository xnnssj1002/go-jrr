package jrr

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type HandlerFunc func(c *Context)

type H map[string]interface{}

type Engine struct {
	// 由于Engine从某种意义上继承了RouterGroup的所有属性和方法，因为 (*Engine).engine 是指向自己的。
	// 这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由。
	*RouterGroup // 将Engine作为最顶层的分组，也就是说Engine拥有RouterGroup所有的能力

	// 路由相关
	router *router        // 路由器
	group  []*RouterGroup // 存储所有的路由分组

	// 模版相关
	htmlTemplates *template.Template // 将所有的模板加载进内存
	funcMap       template.FuncMap   // 所有自定义模板的渲染函数
}

func New() *Engine {
	// Jrr引擎
	e := &Engine{router: newRouter()}
	// 顶级路由分组
	group := &RouterGroup{engine: e}

	e.RouterGroup = group
	e.group = []*RouterGroup{group}

	return e
}

func (e *Engine) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	// 注册中间件时，已经将中间件注册到对应的路由分组中
	// 这里通过接收请求的URL与所有分组匹配，如果匹配到某个分组，就把这个分组注册的中间件添加到本次请求的context的handlers中
	var middlewares []HandlerFunc
	for _, group := range e.group {
		// 本次url的前缀部分含有分组的前缀
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middleware...)
		}
	}

	// 创建本次请求的上下文context
	ctx := NewContext(rsp, req)
	ctx.handlers = middlewares
	// 将框架引擎注入到上下文context中
	ctx.engine = e
	e.router.handle(ctx)
}

// SetFuncMap 设置自定义渲染函数funcMap和加载模板的方法
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// LoadHTMLGlob 加载全局的html模版
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.router.addRoute("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.router.addRoute("POST", pattern, handler)
}

func (e *Engine) Run() error {
	fmt.Printf("server run port: 8888\n")
	return http.ListenAndServe(":8888", e)
}
