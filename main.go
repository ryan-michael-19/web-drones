package main

import (
	"colony-bots/api"
	"colony-bots/impl"
	"colony-bots/schemas"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"colony-bots/webdrones/public/model"
	. "colony-bots/webdrones/public/table"

	. "github.com/go-jet/jet/v2/postgres"

	"github.com/gorilla/sessions"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Use error wrapping???
type AuthError struct {
	originalError error
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("AUTHENTICATION: %s", e.originalError.Error())
}

var sessionStore = sessions.NewCookieStore([]byte("Super secure plz no hax")) // TODO: SET UP ENCRYPTION KEYS
func AuthMiddleWare(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		session, err := sessionStore.Get(r, "SESSION")
		if err != nil {
			return "Authentication Error", &AuthError{originalError: err}
		}
		if operationID == "PostLogin" {
			// check against db
			username, password, ok := r.BasicAuth()
			if !ok {
				return "Authentication Error", &AuthError{originalError: errors.New("AUTHENTICATION: Basic Auth Header issue")}
			}
			stmt := SELECT(Users.Password).FROM(Users).WHERE(Users.Username.EQ(String(username)))
			var hashedPassword model.Users
			err := stmt.Query(schemas.OpenDB(), &hashedPassword)
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.Password), []byte(password))
			if err != nil {
				if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
					return "Authentication Error", &AuthError{originalError: errors.New("Username and password mismatch")}
				} else {
					return "Authentication Error", &AuthError{originalError: err}
				}
			}
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
			// Write cookie into session
			err = session.Save(r, w)
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
		} else if operationID == "PostNewUser" {
			username, password, ok := r.BasicAuth()
			if !ok {
				return "Authentication Error", &AuthError{originalError: errors.New("AUTHENTICATION: Basic Auth Header issue")}
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
			// add new user to db
			stmt := Users.INSERT(
				Users.CreatedAt, Users.UpdatedAt, Users.Username, Users.Password,
			).VALUES(
				NOW(), NOW(), String(username), String(string(hashedPassword)),
			)
			_, err = stmt.Exec(schemas.OpenDB())
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
			// Write cookie into session
			err = session.Save(r, w)
			if err != nil {
				return "Authentication Error", &AuthError{originalError: err}
			}
		} else if session.IsNew {
			// We should have an existing session if we are not logging in or creating a new user
			return "Authentication Error", &AuthError{originalError: errors.New("Must use cookie to access this resource")}
		}
		ctx = context.WithValue(ctx, impl.USERNAME_VALUE, session.Values["username"])
		return f(ctx, w, r, request)
	}

}

func main() {
	RUN_TYPE := os.Args[1]

	if RUN_TYPE == "SERVER" {
		m := []nethttp.StrictHTTPMiddlewareFunc{AuthMiddleWare}
		server := impl.NewServer()
		// create a type that satisfies the `api.StrictServerInterface`, which contains an implementation of every operation from the generated code
		// i := api.NewStrictHandler(server, m)
		i := api.NewStrictHandlerWithOptions(server, m, api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				// All errors returned by nethttp.StrictHTTPMiddlewareFunc middleware are considered response errors,
				// even though bad auth is technically a bad request. (see codcgened stricthandler middleware for details)
				// We're going to be a little messy here and return client error codes with the response error options
				// so we don't have to rewrite the auth middleware.
				var authErr *AuthError
				if errors.As(err, &authErr) {
					http.Error(w, err.Error(), http.StatusUnauthorized)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
		})
		r := http.NewServeMux()
		// get an `http.Handler` that we can use
		h := api.HandlerFromMux(i, r)

		s := &http.Server{
			Handler: h,
			Addr:    "0.0.0.0:8080",
		}
		// And we serve HTTP until the world ends.
		log.Fatal(s.ListenAndServe())
	} else if RUN_TYPE == "SCHEMA" {
		file, err := os.ReadFile("./schemas/schemas.sql")
		if err != nil {
			log.Fatal(err)
		}
		_, err = schemas.OpenDB().Exec(string(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
