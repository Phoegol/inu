package inu

type HandlerInterceptor interface {
	PreHandle(c *Context) bool
	PostHandle(c *Context) error
	AfterCompletion(c *Context) error
}
