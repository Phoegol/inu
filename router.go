package inu

import (
	"fmt"
	"net/http"
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
	Handler func(c *Context) interface{}

	Router struct {
		prefix           []string
		trees            map[string]*Tree
		interceptor      []InterceptorType
		notFound         Handler
		methodNotAllowed Handler
		panic            func(c *Context, err interface{})
	}
)

func New(prefix ...string) *Router {
	if len(prefix) > 0 {
		prefixes := make([]string, len(prefix), len(prefix))
		for _, pre := range prefix {
			if pre := strings.TrimSpace(pre); pre != "" {
				if !strings.HasPrefix(pre, "/") {
					pre = "/" + pre
				}
				if !strings.HasSuffix(pre, "/") {
					pre = strings.TrimSuffix(pre, "/")
				}
				prefixes = append(prefixes, pre)
			}
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

func (r *Router) Use(interceptors ...HandlerInterceptor) {
	for _, interceptor := range interceptors {
		r.interceptor = append(r.interceptor, generateInterceptor(interceptor))
	}
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

	/*if len(r.prefix) > 0 {
		for _, pre := range r.prefix {
			tree.Add(pre+path, handle, interceptor)
		}
	} else {
		tree.Add(path, handle, interceptor)
	}*/
	arr := make([]InterceptorType, len(interceptors))
	for _, interceptor := range interceptors {
		arr = append(arr, generateInterceptor(interceptor))
	}
	tree.Add(path, handle, arr)
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
		http.NotFound(c.W, c.req)
		return
	}
	r.notFound(c)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if r.panic != nil {
		defer func() {
			if err := recover(); err != nil {
				r.panic(&Context{W: w, req: req}, err)
			}
		}()
	}
	if _, ok := r.trees[req.Method]; !ok {
		r.methodNotAllowed(&Context{W: w, req: req})
		return
	}
	if len(r.prefix) > 0 {
		notMatch := true
		for _, pre := range r.prefix {
			if strings.HasPrefix(path, pre) {
				path = strings.TrimPrefix(path, pre)
			}
			notMatch = false
			break
		}
		if notMatch {
			r.notFoundHandle(&Context{W: w, req: req})
			return
		}
	}
	if nd, param := r.trees[req.Method].Find(path, false); nd == nil {
		r.notFoundHandle(&Context{W: w, req: req})
		return
	} else {
		cxt := &Context{W: w, req: req, pathVar: param}

		// 最外部拦截器
		for _, inp := range r.interceptor {
			if !inp.preHandle(cxt) {
				return
			}
		}
		// 路由绑定拦截器
		for _, inp := range nd.interceptor {
			if !inp.preHandle(cxt) {
				return
			}
		}

		res := nd.handle(&Context{W: w, req: req, pathVar: param})
		fmt.Println(res)

		// TODO 环绕方法，返回结果渲染
	}
}

func (r *Router) Run(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), r)
}
