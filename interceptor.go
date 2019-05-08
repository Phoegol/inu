package inu

type HandlerInterceptor func() (func(c *Context) bool, func(c *Context) error, func(c *Context) error)

type InterceptorType struct {
	preHandle       func(c *Context) bool
	postHandle      func(c *Context) error
	afterCompletion func(c *Context) error
}

func generateInterceptor(i HandlerInterceptor) InterceptorType {
	pre, post, after := i()
	return InterceptorType{preHandle: pre, postHandle: post, afterCompletion: after}
}
