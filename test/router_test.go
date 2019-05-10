package test

import (
	"fmt"
	"github.com/cheivin/inu"
	"log"
	"testing"
)

func TestRouter_Run(t *testing.T) {
	r := inu.New("/", "/asd", "")
	/*	if err:=http.ListenAndServe(":8081",r);err!=nil{
		log.Fatal(err)
	}*/
	r.Use(&InterceptorA{}, &InterceptorB{}, &InterceptorC{})
	r.Static("/", "./")
	r.Static("/bbb", "./")
	r.GET("/ping", func(c *inu.Context) (interface{}, inu.RenderType) {
		return `{"name":"BeJson","page":88,"isNonProfit":true}`, inu.Json
	}, &InterceptorD{})

	if err := r.Run(9001); err != nil {
		log.Fatal(err)
	}
}

type InterceptorA struct {
}

func (i *InterceptorA) PreHandle(c *inu.Context) bool {
	fmt.Println("pre 1111")
	return true
}
func (i *InterceptorA) PostHandle(c *inu.Context) error {
	fmt.Println("post 1111")
	return nil
}
func (i *InterceptorA) AfterCompletion(c *inu.Context) error {
	fmt.Println("after 1111")
	return nil
}

type InterceptorB struct {
	inu.HandlerInterceptor
}

func (i *InterceptorB) PreHandle(c *inu.Context) bool {
	fmt.Println("pre 2222")
	return true
}
func (i *InterceptorB) PostHandle(c *inu.Context) error {
	fmt.Println("post 2222")
	return nil
}
func (i *InterceptorB) AfterCompletion(c *inu.Context) error {
	fmt.Println("after 2222")
	return nil
}

type InterceptorC struct {
	inu.HandlerInterceptor
}

func (i *InterceptorC) PreHandle(c *inu.Context) bool {
	fmt.Println("pre 3333")
	return true
}
func (i *InterceptorC) PostHandle(c *inu.Context) error {
	fmt.Println("post 3333")
	return nil
}
func (i *InterceptorC) AfterCompletion(c *inu.Context) error {
	fmt.Println("after 3333")
	return nil
}

type InterceptorD struct {
	inu.HandlerInterceptor
}

func (i *InterceptorD) PreHandle(c *inu.Context) bool {
	fmt.Println("pre 4444")
	panic("Asdasdasd")
	return true
}
func (i *InterceptorD) PostHandle(c *inu.Context) error {
	fmt.Println("post 4444")
	return nil
}
func (i *InterceptorD) AfterCompletion(c *inu.Context) error {
	fmt.Println("after 4444")
	return nil
}
