package inu

import (
	"fmt"
	"net/http"
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
	Context struct {
		w       http.ResponseWriter
		req     *http.Request
		pathVar map[string]string
	}
	Handler func(c *Context) interface{}

	HandlerInterceptor interface {
		preHandle(c *Context) bool
		postHandle(c *Context)
		afterCompletion(c *Context)
	}

	Router struct {
		prefix           []string
		trees            map[string]*Tree
		interceptor      []HandlerInterceptor
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

func (r *Router) Handle(method string, path string, handle Handler, interceptor ...HandlerInterceptor) {
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
	tree.Add(path, handle, interceptor)
}

func (r *Router) GET(path string, handle Handler) {
	r.Handle(http.MethodGet, path, handle)
}

func (r *Router) POST(path string, handle Handler) {
	r.Handle(http.MethodPost, path, handle)
}

func (r *Router) DELETE(path string, handle Handler) {
	r.Handle(http.MethodDelete, path, handle)
}

func (r *Router) PUT(path string, handle Handler) {
	r.Handle(http.MethodPut, path, handle)
}

func (r *Router) PATCH(path string, handle Handler) {
	r.Handle(http.MethodPatch, path, handle)
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
		http.NotFound(c.w, c.req)
		return
	}
	r.notFound(c)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if r.panic != nil {
		defer func() {
			if err := recover(); err != nil {
				r.panic(&Context{w: w, req: req}, err)
			}
		}()
	}
	if _, ok := r.trees[req.Method]; !ok {
		r.methodNotAllowed(&Context{w: w, req: req})
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
			r.notFoundHandle(&Context{w: w, req: req})
			return
		}
	}
	if nd, param := r.trees[req.Method].Find(path, false); nd == nil {
		r.notFoundHandle(&Context{w: w, req: req})
		return
	} else {
		// TODO interceptor环绕
		nd.handle(&Context{w: w, req: req, pathVar: param})
	}
}
