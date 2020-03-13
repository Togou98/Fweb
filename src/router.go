package fw

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// Only one * is allowed
func parsePattern(path string) []string {
	vs := strings.Split(path, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (this *router) addRoute(method string, path string, handler HandlerFunc) {
	parts := parsePattern(path)

	key := method + "-" + path
	_, ok := this.roots[method]
	if !ok {
		this.roots[method] = &node{}
	}
	this.roots[method].insert(path, parts, 0)
	this.handlers[key] = handler
}

func (this *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := this.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.path)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (this *router) getRoutes(method string) []*node {
	root, ok := this.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (this *router) handle(c *Context) {
	n, params := this.getRoute(c.Method, c.Path)

	if n != nil {
		key := c.Method + "-" + n.path
		c.Params = params
		c.handlers = append(c.handlers, this.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
