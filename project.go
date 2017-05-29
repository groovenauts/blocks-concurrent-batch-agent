package pipeline

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

type Project struct {
	ID        string    `json:"id"   datastore:"-"`
	Name      string    `json:"name" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var ErrNoSuchProject = errors.New("No such data in Projects")

func (m *Project) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	if err != nil {
		return err
	}
	return nil
}

func CreateProject(ctx context.Context, proj *Project) (error) {
	t := time.Now()
	proj.CreatedAt = t
	proj.UpdatedAt = t

	err := proj.Validate()
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Projects", nil)
	res, err := datastore.Put(ctx, key, &proj)
	if err != nil {
		log.Errorf(ctx, "@CreateProject %v project: %v\n", err, &proj)
		return err
	}
	// log.Debugf(ctx, "CreateProject res: %v\n", res)
	proj.ID = res.Encode()
	// log.Debugf(ctx, "CreateProject result: %v\n", proj)
	return nil
}


func FindProject(ctx context.Context, id string) (*Project, error) {
	// log.Debugf(ctx, "@FindProject id: %q\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "@FindProject %v id: %q\n", err, id)
		return nil, err
	}
	proj := &Project{ID: id}
	err = datastore.Get(ctx, key, proj)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchProject
	case err != nil:
		log.Errorf(ctx, "@FindProject %v id: %q\n", err, id)
		return nil, err
	}
	return proj, nil
}


func GetAllProject(ctx context.Context) ([]*Project, error) {
	q := datastore.NewQuery("Projects")
	iter := q.Run(ctx)
	var res = []*Project{}
	for {
		m := Project{}
		key, err := iter.Next(&m)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		m.ID = key.Encode()
		res = append(res, &m)
	}
	return res, nil
}

func (m *Project) Key() (*datastore.Key, error) {
	return datastore.DecodeKey(m.ID)
}

func (m *Project) update(ctx context.Context) error {
	t := time.Now()
	m.UpdatedAt = t

	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := m.Key()
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Project) destroy(ctx context.Context) error {
	key, err := m.Key()
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}
