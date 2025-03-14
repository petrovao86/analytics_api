package events

import (
	"time"

	"github.com/uptrace/go-clickhouse/ch"
)

type ApiEvent struct {
	Dt     time.Time `validate:"required"`
	Event  string    `validate:"required"`
	UserId string    `validate:"required"`
	Screen string
	Elem   string
	Amount int
}

type ClickhouseEvent struct {
	ch.CHModel `ch:"table:demo_events_buff"`

	Dt     time.Time `ch:"dt,      type:DateTime,default:now()"`
	Event  string    `ch:"event,   type:LowCardinality(String)"`
	UserId string    `ch:"user_id, type:String"`
	Screen string    `ch:"screen,  type:String"`
	Elem   string    `ch:"elem,    type:String"`
	Amount int       `ch:"amount,  type:Int64"`
}

func (e *ClickhouseEvent) Unmarshal(event *ApiEvent) error {
	e.Dt = event.Dt
	e.Event = event.Event
	e.UserId = event.UserId
	e.Screen = event.Screen
	e.Elem = event.Elem
	e.Amount = event.Amount
	return nil
}
