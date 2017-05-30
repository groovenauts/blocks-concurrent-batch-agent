package pipeline

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type TestSessionStore struct {
	session *sessions.Session
}

func (tss *TestSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	if tss.session == nil {
		tss.session = sessions.NewSession(tss, name)
	}
	return tss.session, nil
}

func (tss *TestSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	tss.session = sessions.NewSession(tss, name)
	return tss.session, nil
}

func (tss *TestSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return nil
}
