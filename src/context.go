package fw

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Container map[string]interface{}

type Context struct {
	// origin objects
	Res http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	// middleware
	handlers []HandlerFunc
	index    int
	// engine pointer
	svc *Svc
}

func newContext(res http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:   req.URL.Path,
		Method: req.Method,
		Req:    req,
		Res: res,
		index:  -1,
	}
}

func (this *Context) Next() {
	this.index++
	s := len(this.handlers)
	for ; this.index < s; this.index++ {
		this.handlers[this.index](this)
	}
}

func (this *Context) Fail(code int, err string) {
	this.index = len(this.handlers)
	this.JSON(code, Container{"message": err})
}

func (this *Context) Param(key string) string {
	value, _ := this.Params[key]
	return value
}

func (this *Context) PostForm(key string) string {
	return this.Req.FormValue(key)
}

func (this *Context) Query(key string) string {
	return this.Req.URL.Query().Get(key)
}

func (this *Context) Status(code int) {
	this.StatusCode = code
	this.Res.WriteHeader(code)
}

func (this *Context) SetHeader(key string, value string) {
	this.Res.Header().Set(key, value)
}

func (this *Context) String(code int, format string, values ...interface{}) {
	this.SetHeader("Content-Type", "text/plain")
	this.Status(code)
	this.Res.Write([]byte(fmt.Sprintf(format, values...)))
}

func (this *Context) JSON(code int, obj interface{}) {
	this.SetHeader("Content-Type", "application/json")
	this.Status(code)
	encoder := json.NewEncoder(this.Res)
	if err := encoder.Encode(obj); err != nil {
		http.Error(this.Res, err.Error(), 500)
	}
}

func (this *Context) Data(code int, data []byte) {
	this.Status(code)
	this.Res.Write(data)
}

// HTML template render
// refer https://golang.org/pkg/html/template/
func (this *Context) HTML(code int, name string, data interface{}) {
	this.SetHeader("Content-Type", "text/html")
	this.Status(code)
	if err := this.svc.htmlTemplates.ExecuteTemplate(this.Res, name, data); err != nil {
		this.Fail(500, err.Error())
	}
}
