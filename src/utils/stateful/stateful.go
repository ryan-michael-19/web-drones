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
		return &AuthError{OriginalError: err}
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
			return &AuthError{OriginalError: err, NewError: errors.New("username already exists")}
		} else {
			return &AuthError{OriginalError: err}
		}
	}
	return nil
}

// Set up an error struct that will log what's going on with the server
// without leaking server errors to the client
// TODO: Use error wrapping???
type AuthError struct {
	OriginalError error
	NewError      error
}

func (e *AuthError) BothErrors() string {
	var newErrorMessage string
	if e.NewError != nil {
		newErrorMessage = e.NewError.Error()
	} else {
		newErrorMessage = ""
	}
	var originalErrorMessage string
	if e.OriginalError != nil {
		originalErrorMessage = e.OriginalError.Error()
	} else {
		originalErrorMessage = ""
	}
	// TODO: Convert this to slog for observability
	return fmt.Sprintf("Authentication error: Original: %s New: %s", originalErrorMessage, newErrorMessage)
}

func (e *AuthError) Error() string {
	if e.NewError == nil {
		return "Unspecified authentication error"
	} else {
		return fmt.Sprintf("Authentication error: %s", e.NewError.Error())
	}
}

func GetDBString() string {
	hostname, present := os.LookupEnv("DB_HOSTNAME")
	if !present || hostname == "" {
		hostname = "localhost"
	}
	dbString := fmt.Sprintf("postgres://user:password@%s:5432/webdrones?sslmode=disable", hostname)
	return dbString
}
