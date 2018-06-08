package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

type (
	Organization struct {
		ID          string    `json:"id" datastore:"-"`
		Name        string    `json:"name" form:"name" validate:"required"`
		Memo        string    `json:"memo" form:"memo"`
		TokenAmount int       `json:"token_amount" form:"token_amount"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
)

func (m *Organization) Validate() error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}

	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Organization) Create(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Organizations", nil)
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()
	return nil
}

func (m *Organization) Destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Organization) Update(ctx context.Context) error {
	m.UpdatedAt = time.Now()

	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Organization) AuthAccessor() *AuthAccessor {
	return &AuthAccessor{Parent: m}
}

func (m *Organization) PipelineAccessor() *PipelineAccessor {
	return &PipelineAccessor{Parent: m}
}

func (m *Organization) GetBackToken(ctx context.Context, pl *Pipeline, handler func() error) error {
	m.TokenAmount = m.TokenAmount + pl.TokenConsumption
	if handler != nil {
		err := handler()
		if err != nil {
			return err
		}
	}
	err := m.Update(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (m *Organization) StartWaitingPipelines(ctx context.Context, handler func(*Pipeline) error) error {
	waitings, err := m.PipelineAccessor().GetWaitings(ctx)
	if err != nil {
		return err
	}

	for _, waiting := range waitings {
		if m.TokenAmount < waiting.TokenConsumption {
			log.Warningf(ctx, "Token %d is shorter than the consumption %d\n", m.TokenAmount, waiting.TokenConsumption)
			return nil
		}
		m.TokenAmount = m.TokenAmount - waiting.TokenConsumption
		waiting.Status = Reserved
		err := waiting.Update(ctx)
		if err != nil {
			return err
		}
		if handler != nil {
			err := handler(waiting)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
