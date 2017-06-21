package admin

import (
	"bytes"
	"html/template"

	"github.com/labstack/echo"
)

type Views interface {
	Render(c echo.Context, code int, name string, data interface{}) error
}

type HandlerViews struct {
	templates *template.Template
}

func (t *HandlerViews) Render(c echo.Context, code int, name string, data interface{}) (err error) {
	buf := new(bytes.Buffer)
	if err = t.templates.ExecuteTemplate(buf, name, data); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}
