package events

import (
	"context"
	"fmt"
	"net/url"

	"github.com/uptrace/go-clickhouse/ch"
)

type IRepository interface {
	Insert(*ApiEvent) error
}

func BuildRepo(addr string) (IRepository, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "ch", "clickhouse":
		return NewClickhouse(addr)
	}
	return nil, fmt.Errorf("unknown evens repository type %s", u.Scheme)
}

type clickhouseRepo struct {
	db *ch.DB
}

func NewClickhouse(addr string) (IRepository, error) {
	return &clickhouseRepo{
		db: ch.Connect(
			ch.WithDSN(addr),
			ch.WithMaxRetries(3),
		),
	}, nil
}

func (c *clickhouseRepo) Insert(e *ApiEvent) error {
	chE := new(ClickhouseEvent)
	if err := chE.Unmarshal(e); err != nil {
		return err
	}
	_, err := c.db.NewInsert().Model(chE).Exec(context.TODO())
	return err
}
