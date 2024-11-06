package utils

import (
	"errors"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ryan-michael-19/web-drones/schemas"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"
	"golang.org/x/crypto/bcrypt"
)

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
	_, err = stmt.Exec(schemas.OpenDB())
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
