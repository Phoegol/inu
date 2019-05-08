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
	Handler func(c *Context) (interface{}, RenderType)

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
	var arr []InterceptorType
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
		http.NotFound(c.w, c.Req)
		return
	}
	r.notFound(c)
}

func (r *Router) panicHandle(c *Context, err interface{}) {
	if r.notFound == nil {
		http.Error(c.w, fmt.Sprintf("failed to parse events: %v", err), http.StatusInternalServerError)
		return
	}
	r.panic(c, err)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	header := w.Header()
	cxt := &Context{w: w, Req: req, header: &header}
	serve := false

	defer func() {
		if rc := recover(); rc != nil {
			r.panicHandle(cxt, rc)
			return
		}
	}()

	if _, ok := r.trees[req.Method]; !ok {
		r.methodNotAllowed(cxt)
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
			r.notFoundHandle(cxt)
			return
		}
	}
	if nd, param := r.trees[req.Method].Find(path, false); nd == nil {
		r.notFoundHandle(cxt)
		return
	} else {
		cxt.PathVar = param
		serve = true
		var itcpts []InterceptorType

		// 最外部拦截器
		for _, inp := range r.interceptor {
			if !inp.preHandle(cxt) {
				serve = false
				break
			}
			itcpts = append([]InterceptorType{inp}, itcpts...)
		}
		if serve {
			// 路由绑定拦截器
			for _, inp := range nd.interceptor {
				if !inp.preHandle(cxt) {
					serve = false
					return
				}
				itcpts = append([]InterceptorType{inp}, itcpts...)
			}
		}
		if serve {
			res, renderType := nd.handle(cxt)
			for _, itcpt := range itcpts {
				if err := itcpt.postHandle(cxt); err != nil {
					r.panicHandle(cxt, err)
					return
				}
			}
			if res != nil {
				render(cxt.w, res, renderType)
			}
		}
		for _, itcpt := range itcpts {
			if err := itcpt.afterCompletion(cxt); err != nil {
				r.panicHandle(cxt, err)
				return
			}
		}
	}

	// TODO afterCompletion待处理，write逻辑待处理

	// TODO cookie和session待处理
}

func (r *Router) Run(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), r)
}
