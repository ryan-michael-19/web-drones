package schemas

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenDB() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), "postgres://gorm:gorm@localhost:5432/gorm")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

// TODO: Get rid of these??
type Metadata struct {
	CreatedAt time.Time `db:"-"`
	UpdatedAt time.Time `db:"-"`
	DeletedAt time.Time `db:"-"`
}

type Bots struct {
	Metadata
	ID         int `db:"-"`
	Identifier string
	Name       string
}

type BotActions struct {
	Metadata
	ID                  int `db:"-"`
	Bot_ID              int `db:"-"`
	Time_Action_Started time.Time
	New_X               float64
	New_Y               float64
}

type Mines struct {
	Metadata
	ID int `db:"-"`
	X  float64
	Y  float64
}
