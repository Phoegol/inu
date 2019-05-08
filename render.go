package inu

import "net/http"

type RenderType int

const (
	String RenderType = iota
	Html
	Json
	Jsonp
	Xml
)

var (
	stringContentType = "text/plain; charset=utf-8"
	jsonContentType   = "application/json; charset=utf-8"
	jsonpContentType  = "application/javascript; charset=utf-8"
)

func render(w http.ResponseWriter, obj interface{}, renderType RenderType) error {
	var err error
	switch renderType {
	case Html:
		break
	case Json:
		renderJson(w, obj)
		break
	case Jsonp:
		break
	case Xml:
		break
	case String:
		renderString(w, obj)
	default:
		renderString(w, obj)
	}
	return err
}

func writeContentType(w http.ResponseWriter, value string) {
	header := w.Header()
	header.Add("Content-Type", value)
}

func renderString(w http.ResponseWriter, obj interface{}) (err error) {
	writeContentType(w, stringContentType)
	s := obj.(string)
	_, err = w.Write([]byte(s))
	return
}

func renderJson(w http.ResponseWriter, obj interface{}) (err error) {
	writeContentType(w, jsonContentType)
	s := obj.(string)
	_, err = w.Write([]byte(s))
	return
}
