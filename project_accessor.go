package pipeline

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type ProjectAccessor struct {
}

var ErrNoSuchProject = errors.New("No such data in Projects")

func (acc *ProjectAccessor)Create(ctx context.Context, m *Project) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Project", nil)
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()
	return nil
}

func (acc *ProjectAccessor)Find(ctx context.Context, id string) (*Project, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "Failed to decode id(%v) to key because of %v \n", id, err)
		return nil, err
	}
	m := &Project{ID: id}
	err = datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchProject
	case err != nil:
		log.Errorf(ctx, "Failed to Get pipeline key(%v) to key because of %v \n", key, err)
		return nil, err
	}
	return m, nil
}

func (acc *ProjectAccessor)GetAll(ctx context.Context) ([]*Project, error) {
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
