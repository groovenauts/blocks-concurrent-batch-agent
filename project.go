package pipeline

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Project struct {
	ID   string `datastore:"-"`
	Name string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

var ErrNoSuchProject = errors.New("No such data in Projects")

func CreateProject(ctx context.Context, proj *Project) (error) {
	key := datastore.NewIncompleteKey(ctx, "Projects", nil)
	t := time.Now()
	proj.CreatedAt = t
	proj.UpdatedAt = t
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
	key, err := m.Key()
	if err != nil {
		return err
	}
	t := time.Now()
	m.UpdatedAt = t
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
