package inu

import (
	"log"
	"testing"
)

func TestRouter_Run(t *testing.T) {
	r := New()
	/*	if err:=http.ListenAndServe(":8081",r);err!=nil{
		log.Fatal(err)
	}*/
	r.GET("/ping", func(c *Context) interface{} {
		if i, err := c.W.Write([]byte("pong")); err != nil {
			log.Fatal(i, err)
		} else {
			log.Println(i)
		}
		return "OK"
	})

	if err := r.Run(8081); err != nil {
		log.Fatal(err)
	}
}
