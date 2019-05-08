package inu

import "net/http"

type Context struct {
	W       http.ResponseWriter
	req     *http.Request
	pathVar map[string]string
}
