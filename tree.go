package inu

import (
	"regexp"
	"strings"
)

type (
	Tree struct {
		root *Node
	}

	Node struct {
		key           string
		value         string
		path          string
		children      []*Node
		regexChildren []*Node
		regex         RegexInfo
	}

	RegexInfo struct {
		name  string
		regex *regexp.Regexp
	}

	NodeValue struct {
	}
)

func NewNode(key string) *Node {
	return &Node{
		key:      key,
		children: []*Node{},
	}
}

func NewTree() *Tree {
	return &Tree{
		root: NewNode("/"),
	}
}

func (t *Tree) Add(pattern string, value string) {
	var currentNode = t.root

	if pattern != currentNode.key {
		pattern = strings.TrimPrefix(pattern, "/")
		nodKeys := strings.Split(pattern, "/")
	l:
		for _, key := range nodKeys {
			for _, node := range currentNode.children {
				if node.key == key {
					currentNode = node
					continue l
				}
			}
			node := NewNode(key)
			currentNode.children = append(currentNode.children, node)
			currentNode = node
		}
	}
	currentNode.value = value
}

func (t *Tree) Find(pattern string) *Node {
	var currentNode = t.root
	if pattern != currentNode.key {
		pattern = strings.TrimPrefix(pattern, "/")
		nodKeys := strings.Split(pattern, "/")
	l:
		for _, key := range nodKeys {
			for _, node := range currentNode.children {
				if node.key == key {
					currentNode = node
					continue l
				}
			}
			return nil
		}
	}
	return currentNode
}

func fmtRegex(str string) *RegexInfo {
	if !strings.HasPrefix(str, "{") || !strings.HasSuffix(str, "}") {
		return nil
	}
	str = strings.TrimSuffix(strings.TrimSuffix(str, "{"), "}")
	spIdx := strings.IndexAny(str, ":")
	switch spIdx {
	case -1:
		return &RegexInfo{name: str}
	case len(str) - 1:
		return &RegexInfo{name: str[:len(str)-1]}
	default:
		reg := strings.Split(str, ":")
		if r, err := regexp.Compile(reg[1]); err != nil {
			panic("url regexp err")
		} else {
			return &RegexInfo{name: reg[0], regex: r}
		}
	}
}
