package pipeline

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
)

type Flash struct {
	store   sessions.Store
	Alerts  []interface{}
	Notices []interface{}
}

func (f *Flash) session(c echo.Context) (*sessions.Session, error) {
	session, err := f.store.Get(c.Request(), "admin-session")
	return session, err
}

func (f *Flash) set(c echo.Context, name, value string) error {
	session, err := f.session(c)
	if err != nil {
		return err
	}
	session.AddFlash(name, value)
	return nil
}

func (f *Flash) load(c echo.Context) error {
	session, err := f.session(c)
	if err != nil {
		return err
	}
	f.Alerts = session.Flashes("alert")
	f.Notices = session.Flashes("notice")
	return nil
}
