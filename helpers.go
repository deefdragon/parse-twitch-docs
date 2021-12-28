package main

import (
	"strings"

	"golang.org/x/net/html"
)

func nodeHasClass(node *html.Node, class string) bool {

	for _, v := range node.Attr {
		classes := strings.Split(v.Val, " ")
		if v.Key == "class" {
			for _, c := range classes {
				if c == class {
					return true
				}
			}
		}
	}

	return false
}

func findChildNodeWithClass(node *html.Node, class string) *html.Node {

	n := node.FirstChild
	for {

		if n == nil {
			return nil
		}

		if nodeHasClass(n, class) {
			return n
		}
		n = n.NextSibling
	}

}

func getNodeChildren(node *html.Node) []*html.Node {
	nodes := make([]*html.Node, 0)

	n := node.FirstChild
	for {
		if n == nil {
			return nodes
		}
		nodes = append(nodes, n)
		n = n.NextSibling
	}
}

func filter[T any](nodes []T, keepFunc func(T) (keep bool)) []T {
	out := make([]T, 0)
	for _, v := range nodes {
		if keepFunc(v) {
			out = append(out, v)
		}
	}
	return out
}
