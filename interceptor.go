package inu

type HandlerInterceptor func() (InterceptorPreHandle, InterceptorPostHandle, InterceptorAfterCompletion)

type InterceptorPreHandle func(c *Context) bool
type InterceptorPostHandle func(c *Context) error
type InterceptorAfterCompletion func(c *Context) error

type InterceptorType struct {
	preHandle       InterceptorPreHandle
	postHandle      InterceptorPostHandle
	afterCompletion InterceptorAfterCompletion
}

func generateInterceptor(i HandlerInterceptor) InterceptorType {
	pre, post, after := i()
	return InterceptorType{preHandle: pre, postHandle: post, afterCompletion: after}
}
