package pipeline

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
)

type Flash struct {
	session *sessions.Session
	Alerts  []interface{}
	Notices []interface{}
}

func (f *Flash) set(c echo.Context, name, value string) error {
	f.session.AddFlash(name, value)
	return nil
}

func (f *Flash) load(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	log.Debugf(ctx, "Flash#load session: %v\n", f.session)

	f.Alerts = f.session.Flashes("alert")
	f.Notices = f.session.Flashes("notice")
	return nil
}
