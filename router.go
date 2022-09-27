package jrr

import (
	"fmt"
	"strings"
)

type router struct {
	roots    map[string]*node       // 存储每种请求方式的Trie 树根节点
	handlers map[string]HandlerFunc // 存储每种请求方式的 HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, v := range vs {
		if v != "" {
			parts = append(parts, v)
			if v[0] == '*' {
				break
			}
		}
	}

	return parts
}

// addRoute 添加路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	fmt.Printf("add route %4s - %s\n", method, pattern)
	parts := parsePattern(pattern)

	key := method + "-" + pattern

	// 判断当前请求方法的前缀树是否存在
	if _, exists := r.roots[method]; !exists {
		r.roots[method] = &node{} // 前缀树的根结点为空node
	}

	r.roots[method].insert(pattern, parts, 0)

	r.handlers[key] = handler
}

// getRoute 获取路由  根据请求的真实路由，找到注册时的路由
// 如果时动态路由的话，根据真实路由: /a/b/10 -> 找到注册的路由：/a/b/:id
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	// 请求方法树是否存在
	root, exists := r.roots[method]
	if !exists {
		return nil, nil
	}

	params := make(map[string]string)
	parts := parsePattern(path)

	n := root.search(parts, 0)
	if n != nil {
		findParts := parsePattern(n.pattern) // 需要将动态路由的值绑定到指定参数上
		for index, findPart := range findParts {
			// findPart: /a/b/:id ; patch: /a/b/10 => params["id"] = 10
			if findPart[0] == ':' {
				params[findPart[1:]] = parts[index]
			}
			// findPart /a/*filePath ; patch: /a/b/c/d.css => params["filePath"] = "b/c/d.css"
			if findPart[0] == '*' {
				params[findPart[1:]] = strings.Join(parts[index:], "/")
			}
		}
		return n, params
	}

	return nil, params
}

// handle 根据请求上下文，处理请求，并响应
func (r *router) handle(ctx *Context) {
	// 通过trie树获取对应的handler
	n, params := r.getRoute(ctx.Method, ctx.Path)
	if n != nil {
		ctx.Params = params
		key := ctx.Method + "-" + n.pattern
		// 将找到handle函数添加到context的handlers中
		ctx.handlers = append(ctx.handlers, r.handlers[key])
		//r.handlers[key](ctx)  // 这里不再执行，有context同步调度执行middleware与业务handler
	} else {
		ctx.handlers = append(ctx.handlers, func(c *Context) {
			_, _ = c.Writer.Write([]byte(fmt.Sprintf("404 NOT FOUND %s\n", ctx.Req.URL)))
		})
	}

	// context中所有的middleware与业务handler都已经注册完毕，通过context调度执行自身里面所有的handlers
	ctx.Run()

}
