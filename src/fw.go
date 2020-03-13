package fw

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)
type (
	RouterGroup struct {
		prefix      string
		middlewares []HandlerFunc // support middleware
		parent      *RouterGroup  // support nesting
		svc      *Svc       // all groups share a Engine instance
	}

	Svc struct {
		*RouterGroup
		router        *router
		groups        []*RouterGroup     // store all groups
		htmlTemplates *template.Template // for html render
		funcMap       template.FuncMap   // for html render
	}
)

func New() *Svc {
	svc := &Svc{router: newRouter()}
	svc.RouterGroup = &RouterGroup{svc:svc}
	svc.groups = []*RouterGroup{svc.RouterGroup}
	return svc
}


// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (this *RouterGroup) Group(prefix string) *RouterGroup {
	svc := this.svc
	newGroup := &RouterGroup{
		prefix: this.prefix + prefix,
		parent: this,
		svc:    svc,
	}
	svc.groups = append(svc.groups, newGroup)
	return newGroup
}

func (this *RouterGroup) Use(middlewares ...HandlerFunc) {
	this.middlewares = append(this.middlewares, middlewares...)
}

func (this *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := this.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	this.svc.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (this *RouterGroup) GET(path string, handler HandlerFunc) {
	this.addRoute("GET", path, handler)
}

// POST defines the method to add POST request
func (this *RouterGroup) POST(path string, handler HandlerFunc) {
	this.addRoute("POST", path, handler)
}

// create static handler
func (this *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(this.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Res, c.Req)
	}
}

// serve static files
func (this *RouterGroup) Static(relativePath string, root string) {
	handler := this.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	this.GET(urlPattern, handler)
}

// for custom render function
func (this *Svc) SetFuncMap(funcMap template.FuncMap) {
	this.funcMap = funcMap
}

func (this *Svc) LoadHTMLGlob(pattern string) {
	this.htmlTemplates = template.Must(template.New("").Funcs(this.funcMap).ParseGlob(pattern))
}

// Run defines the method to start a http server
func (this *Svc) Run(addr string) (err error) {
	return http.ListenAndServe(addr, this)
}

func (this *Svc) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range this.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(res, req)
	c.handlers = middlewares
	c.svc = this
	this.router.handle(c)
}
