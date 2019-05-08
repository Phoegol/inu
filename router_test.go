package inu

import (
	"fmt"
	"log"
	"testing"
)

func TestRouter_Run(t *testing.T) {
	r := New()
	/*	if err:=http.ListenAndServe(":8081",r);err!=nil{
		log.Fatal(err)
	}*/
	r.Use(aaa, bbb, ccc)
	r.GET("/ping", func(c *Context) (interface{}, RenderType) {
		c.header.Add("name", "cheivin")
		return `{"name":"BeJson","page":88,"isNonProfit":true}`, Json
	}, ddd)

	if err := r.Run(9001); err != nil {
		log.Fatal(err)
	}
}

func aaa() (InterceptorPreHandle, InterceptorPostHandle, InterceptorAfterCompletion) {
	return func(c *Context) bool {
			fmt.Println("pre 1111")
			return true
		}, func(c *Context) error {
			fmt.Println("post 1111")
			return nil
		}, func(c *Context) error {
			fmt.Println("after 1111")
			return nil
		}
}

func bbb() (InterceptorPreHandle, InterceptorPostHandle, InterceptorAfterCompletion) {
	return func(c *Context) bool {
			fmt.Println("pre 2222")
			return true
		}, func(c *Context) error {
			fmt.Println("post 2222")
			return nil
		}, func(c *Context) error {
			fmt.Println("after 2222")
			return nil
		}
}

func ccc() (InterceptorPreHandle, InterceptorPostHandle, InterceptorAfterCompletion) {
	return func(c *Context) bool {
			fmt.Println("pre 3333")
			return true
		}, func(c *Context) error {
			fmt.Println("post 3333")
			return nil
		}, func(c *Context) error {
			fmt.Println("after 3333")
			return nil
		}
}

func ddd() (InterceptorPreHandle, InterceptorPostHandle, InterceptorAfterCompletion) {
	return func(c *Context) bool {
			fmt.Println("pre 4444")
			return true
		}, func(c *Context) error {
			fmt.Println("post 4444")
			return nil
		}, func(c *Context) error {
			fmt.Println("after 4444")
			return nil
		}
}
