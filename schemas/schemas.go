package schemas

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenDB(ctx context.Context) *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), "postgres://gorm:gorm@localhost:5432/gorm")
	if err != nil {
		log.Fatal(err)
	}
	// make sure the database is up and running
	_, err = conn.Exec(ctx, ";")
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
	ID             int `db:"-"`
	Identifier     string
	Name           string
	InventoryCount int
}

type BotActions struct {
	Metadata
	ID                int `db:"-"`
	Bot_Key           int `db:"-"`
	TimeActionStarted time.Time
	New_X             float64
	New_Y             float64
}

type Mines struct {
	Metadata
	ID int `db:"-"`
	X  float64
	Y  float64
}

func BuildInsert(tableName string, colNames ...string) string {
	// TODO: This feels very injectable...
	valueFormats := []string{} // ["$1", "$2", etc..]
	for i := range colNames {
		valueFormats = append(valueFormats, "$"+fmt.Sprintf("%d", i+1))
	}
	valString := "NOW(),NULL,NULL," + strings.Join(valueFormats, ",")
	colString := "created_at,updated_at,deleted_at," + strings.Join(colNames, ",")
	sqlString := "INSERT INTO " + tableName + "(" + colString + ") VALUES (" + valString + ")"
	return sqlString
}
