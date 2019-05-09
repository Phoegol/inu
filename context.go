package inu

import "net/http"

type Context struct {
	w       http.ResponseWriter
	Req     *http.Request
	PathVar map[string]string
	header  *http.Header
}
