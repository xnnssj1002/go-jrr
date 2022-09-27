package jrr

import (
	"net/http"
	"path"
)

// RouterGroup 路由分组
type RouterGroup struct {
	prefix     string        // group前缀
	middleware []HandlerFunc // 当前分组拥有的中间件列表
	parent     *RouterGroup  // 父级分组
	engine     *Engine       // 所有分组共享一个引擎实例
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	group := &RouterGroup{
		prefix: g.prefix + prefix,
		parent: g,
		engine: g.engine,
	}
	g.engine.group = append(g.engine.group, group)
	return group
}

// Use 为路由分组添加中间件
func (g *RouterGroup) Use(fns ...HandlerFunc) {
	g.middleware = append(g.middleware, fns...)
}

// Static 设置静态文件路径
func (g *RouterGroup) Static(relativePath string, root string) {
	// 解析Go 标准库 http.FileServer 实现静态文件服务，参考路径地址：https://www.jb51.net/article/146035.htm
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	g.GET(urlPattern, handler)
}

// createStaticHandler 创建处理静态文件的handler
func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 获取静态文件的绝对路径
	absolutePath := path.Join(g.prefix, relativePath)

	// func StripPrefix(prefix string, h Handler) Handler
	// StripPrefix的原型。StripPrefix将URL中的前缀中的prefix字符串删除，然后再交给后面的Handler处理，一般是http.FileServer()的返回值。
	// 如果URL不是以prefix开始，或者prefix包含转移字符，最后结果都会返回404，因此要精确匹配URL和prefix

	// func FileServer(root FileSystem) Handler
	// FileServer()被告知静态文件的根是root。
	// 这个函数返回一个Handler，这个Handler向http request提供位于上面代码root变量的文件系统的内容，直接定位到这个目录下的index.html文件。
	// root一般使用http.Dir(“yourFilePath”)
	// 当我们URL是根路径"/“时，http.Handle(”/", http.FileServer(http.Dir(“yourFilePath”))，这个没有什么问题的。
	// 可是当我们把URL随便改成其他的，如“/hello"这样。再访问这个URL时，就会报404
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Params["filepath"]
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}

}

// addRouter 通过路由组添加路由
// 调用g.engine.router.addRouter来实现了路由的映射。
func (g *RouterGroup) addRouter(method string, pattern string, fn HandlerFunc) {
	comPattern := g.prefix + pattern
	g.engine.router.addRoute(method, comPattern, fn)
}

func (g *RouterGroup) GET(pattern string, fn HandlerFunc) {
	g.addRouter("GET", pattern, fn)
}

func (g *RouterGroup) POST(pattern string, fn HandlerFunc) {
	g.addRouter("POST", pattern, fn)
}
