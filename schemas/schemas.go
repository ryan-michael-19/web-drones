package schemas

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func OpenDB() *sql.DB {
	// conn, err := pgxpool.New(context.Background(), "postgres://user:password@localhost:5432/webdrones")
	db, err := sql.Open(
		"pgx", "postgres://user:password@localhost:5432/webdrones",
	)
	if err != nil {
		log.Fatal(err)
	}

	// make sure the database is up and running
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
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
	valString := "NOW(),NOW()," + strings.Join(valueFormats, ",")
	colString := "created_at,updated_at," + strings.Join(colNames, ",")
	sqlString := "INSERT INTO " + tableName + "(" + colString + ") VALUES (" + valString + ")"
	return sqlString
}
