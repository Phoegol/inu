package inu

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	methods = map[string]struct{}{
		http.MethodGet:    {},
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodDelete: {},
		http.MethodPatch:  {},
	}
)

type (
	Handler func(c *Context) (interface{}, RenderType)

	staticDir struct {
		prefix string
		dir    string
		list   bool
	}

	Router struct {
		prefix           []string
		trees            map[string]*Tree
		interceptor      []HandlerInterceptor
		staticDirs       []staticDir
		notFound         Handler
		methodNotAllowed Handler
		panic            func(c *Context, err interface{})
	}
)

func New(prefix ...string) *Router {
	if len(prefix) > 0 {
		prefixes := make([]string, 0)
		hasRoot := false
		for _, pre := range prefix {
			if pre := strings.TrimSpace(pre); pre == "" || pre == "/" {
				hasRoot = true
			} else {
				if !strings.HasPrefix(pre, "/") {
					pre = "/" + pre
				}
				if !strings.HasSuffix(pre, "/") {
					pre = strings.TrimSuffix(pre, "/")
				}
				prefixes = append(prefixes, pre)
			}
		}
		if hasRoot {
			prefixes = append(prefixes, "/")
		}
		return &Router{
			prefix: prefixes,
			trees:  make(map[string]*Tree),
		}
	} else {
		return &Router{
			prefix: []string{},
			trees:  make(map[string]*Tree),
		}
	}
}

func (r *Router) Static(prefix string, dir string) {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "/"
	} else {
		prefix = "/" + prefix + "/"
	}
	r.staticDirs = append(r.staticDirs, staticDir{prefix: prefix, dir: dir, list: false})
}

func (r *Router) StaticDir(prefix string, dir string) {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "/"
	} else {
		prefix = "/" + prefix + "/"
	}
	r.staticDirs = append(r.staticDirs, staticDir{prefix: prefix, dir: dir, list: true})
}

func (r *Router) Use(interceptors ...HandlerInterceptor) {
	/*	for _, interceptor := range interceptors {
		r.interceptor = append(r.interceptor, generateInterceptor(interceptor))
	}*/
	r.interceptor = append(r.interceptor, interceptors...)
}
func (r *Router) Handle(method string, path string, handle Handler, interceptors []HandlerInterceptor) {
	if _, ok := methods[method]; !ok {
		panic(fmt.Errorf("invalid method"))
	}
	tree, ok := r.trees[method]
	if !ok {
		tree = NewTree()
		r.trees[method] = tree
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	/*var arr []InterceptorType
	for _, interceptor := range interceptors {
		arr = append(arr, generateInterceptor(interceptor))
	}*/
	tree.Add(path, handle, interceptors)
}

func (r *Router) GET(path string, handle Handler, interceptor ...HandlerInterceptor) {
	r.Handle(http.MethodGet, path, handle, interceptor)
}

func (r *Router) POST(path string, handle Handler, interceptor ...HandlerInterceptor) {
	r.Handle(http.MethodPost, path, handle, interceptor)
}

func (r *Router) DELETE(path string, handle Handler, interceptor ...HandlerInterceptor) {
	r.Handle(http.MethodDelete, path, handle, interceptor)
}

func (r *Router) PUT(path string, handle Handler, interceptor ...HandlerInterceptor) {
	r.Handle(http.MethodPut, path, handle, interceptor)
}

func (r *Router) PATCH(path string, handle Handler, interceptor ...HandlerInterceptor) {
	r.Handle(http.MethodPatch, path, handle, interceptor)
}

func (r *Router) NotFoundFunc(handler Handler) {
	r.notFound = handler
}

func (r *Router) MethodNotAllowedFunc(handler Handler) {
	r.methodNotAllowed = handler
}

func (r *Router) PanicFunc(handler func(c *Context, err interface{})) {
	r.panic = handler
}

func (r *Router) notFoundHandle(c *Context) {
	if r.notFound == nil {
		http.NotFound(c.w, c.Req)
		return
	}
	r.notFound(c)
}

func (r *Router) panicHandle(c *Context, err interface{}) {
	log.SetPrefix("[inu]")
	log.Println(err)
	if r.notFound == nil {
		http.Error(c.w, fmt.Sprintf("failed to parse events: %v", err), http.StatusInternalServerError)
		return
	}
	r.panic(c, err)
}

func (r *Router) matchFile(path string) string {
	for _, sDir := range r.staticDirs {
		if strings.HasPrefix(path, sDir.prefix) {
			file := sDir.dir + strings.TrimPrefix(path, sDir.prefix)
			if f, err := os.Stat(file); err == nil && (!f.IsDir() || sDir.list) {
				return file
			}
		}
	}
	return ""
}

func (r *Router) staticHandle(w http.ResponseWriter, req *http.Request, path string) {
	http.ServeFile(w, req, path)
}

func (r *Router) routerHandle(cxt *Context, nd *Node) {
	// 判断请求方法
	if _, ok := r.trees[cxt.Req.Method]; !ok {
		r.methodNotAllowed(cxt)
		return
	}
	serve := true
	var itcpts []HandlerInterceptor

	// 最外部拦截器
	for _, inp := range r.interceptor {
		if !inp.PreHandle(cxt) {
			serve = false
			break
		}
		itcpts = append([]HandlerInterceptor{inp}, itcpts...)
	}
	if serve {
		// 路由绑定拦截器
		for _, inp := range nd.interceptor {
			if !inp.PreHandle(cxt) {
				serve = false
				break
			}
			itcpts = append([]HandlerInterceptor{inp}, itcpts...)
		}
	}
	if serve {
		res, renderType := nd.handle(cxt)
		for _, itcpt := range itcpts {
			if err := itcpt.PostHandle(cxt); err != nil {
				r.panicHandle(cxt, err)
				return
			}
		}
		if res != nil {
			if err := render(cxt.w, res, renderType); err != nil {
				r.panicHandle(cxt, err)
				return
			}
		}
	}
	for _, itcpt := range itcpts {
		if err := itcpt.AfterCompletion(cxt); err != nil {
			r.panicHandle(cxt, err)
			return
		}
	}

	// TODO afterCompletion待处理，write逻辑待处理

	// TODO cookie和session待处理
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	header := w.Header()
	cxt := &Context{w: w, Req: req, header: &header}

	defer func() {
		// 异常捕获
		if rc := recover(); rc != nil {
			r.panicHandle(cxt, rc)
		}
	}()
	// 匹配前缀
	if len(r.prefix) > 0 {
		notMatch := true
		for _, pre := range r.prefix {
			if strings.HasPrefix(path, pre) {
				if pre != "/" {
					path = strings.TrimPrefix(path, pre)
				}
				notMatch = false
				break
			}
		}
		if notMatch {
			r.notFoundHandle(cxt)
			return
		}
	}
	if file := r.matchFile(path); file != "" { // 静态资源匹配
		r.staticHandle(cxt.w, cxt.Req, file)
		return
	} else if nd, param := r.trees[req.Method].Find(path, false); nd != nil { // 路由匹配
		cxt.PathVar = param
		r.routerHandle(cxt, nd)
		return
	}
	r.notFoundHandle(cxt)
}

func (r *Router) Run(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), r)
}
