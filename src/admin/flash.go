package admin

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/gae_support"
)

type Flash struct {
	Alert  string
	Notice string
}

func setFlash(c echo.Context, name, value string) {
	setFlashWithExpire(c, name, value, time.Now().Add(10*time.Minute))
}

func setFlashWithExpire(c echo.Context, name, value string, expire time.Time) {
	cookie := new(http.Cookie)
	cookie.Path = "/admin/"
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = expire
	c.SetCookie(cookie)
}

func loadFlash(c echo.Context) *Flash {
	f := Flash{}
	cookie, err := c.Cookie("alert")
	if err == nil {
		f.Alert = cookie.Value
	}
	cookie, err = c.Cookie("notice")
	if err == nil {
		f.Notice = cookie.Value
	}
	return &f
}

func clearFlash(c echo.Context) {
	_, err := c.Cookie("alert")
	if err == nil {
		setFlashWithExpire(c, "alert", "", time.Now().AddDate(0, 0, 1))
	}
	_, err = c.Cookie("notice")
	if err == nil {
		setFlashWithExpire(c, "notice", "", time.Now().AddDate(0, 0, 1))
	}
}

func withFlash(impl func(c echo.Context) error) func(c echo.Context) error {
	return gae_support.With(func(c echo.Context) error {
		f := loadFlash(c)
		c.Set("flash", f)
		clearFlash(c)
		return impl(c)
	})
}
