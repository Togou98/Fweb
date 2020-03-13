package fw

import (
	"fmt"
	"strings"
)

type node struct {
	path  string
	part     string
	children []*node
	isWild   bool
}

func (this *node) String() string {
	return fmt.Sprintf("node{path=%s, part=%s, isWild=%t}", this.path, this.part, this.isWild)
}

func (this *node) insert(path string, parts []string, height int) {
	if len(parts) == height {
		this.path = path
		return
	}
	part := parts[height]
	child := this.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		this.children = append(this.children, child)
	}
	child.insert(path, parts, height+1)
}

func (this *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(this.part, "*") {
		if this.path == "" {
			return nil
		}
		return this
	}

	part := parts[height]
	children := this.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this *node) travel(list *([]*node)) {
	if this.path != "" {
		*list = append(*list, this)
	}
	for _, child := range this.children {
		child.travel(list)
	}
}

func (this *node) matchChild(part string) *node {
	for _, child := range this.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

func (this *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range this.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
