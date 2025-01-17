package stateful

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ryan-michael-19/web-drones/utils/stateless"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"
	"golang.org/x/crypto/bcrypt"
)

var _ = BuildLogger()
var DB = OpenDB()

func OpenDB() *sql.DB {
	db, err := sql.Open(
		"pgx", GetDBString(),
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

func BuildLogger() *slog.Logger {
	replace := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "source" {
			src := a.Value.Any().(*slog.Source)
			return slog.String("source", src.File+":"+strconv.Itoa(src.Line))
		}
		return a
	}
	logBase := path.Join(".", "logs")
	err := os.MkdirAll(logBase, os.ModePerm)
	if err != nil {
		log.Fatalf(err.Error())
	}
	logName := path.Join(logBase, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02-15-04-05")))
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf(err.Error())
	}
	w := io.MultiWriter(logFile, os.Stdout)
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replace,
	}))
	slog.SetDefault(logger)
	return logger
}

func CreateNewUser(username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &stateless.AuthError{OriginalError: err}
	}
	// add new user to db
	stmt := Users.INSERT(
		Users.CreatedAt, Users.UpdatedAt, Users.Username, Users.Password,
	).VALUES(
		NOW(), NOW(), String(username), String(string(hashedPassword)),
	)
	_, err = stmt.Exec(DB)
	if err != nil {
		if err.(*pgconn.PgError).Code == "23505" {
			return &stateless.AuthError{OriginalError: err, NewError: errors.New("username already exists")}
		} else {
			return &stateless.AuthError{OriginalError: err}
		}
	}
	return nil
}

func GetDBString() string {
	hostname, present := os.LookupEnv("DB_HOSTNAME")
	if !present || hostname == "" {
		hostname = "localhost"
	}
	dbString := fmt.Sprintf("postgres://user:password@%s:5432/webdrones?sslmode=disable", hostname)
	return dbString
}
