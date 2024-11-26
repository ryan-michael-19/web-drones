package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/impl"
	"github.com/ryan-michael-19/web-drones/utils"

	"github.com/antonlindstrom/pgstore"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"

	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"golang.org/x/crypto/bcrypt"
)

// var sessionStore = sessions.NewCookieStore([]byte("Super secure plz no hax")) // TODO: SET UP ENCRYPTION KEYS
func makeStore() *pgstore.PGStore {
	// TODO: Manage sessions in separate database?
	sessionStore, err := pgstore.NewPGStore(utils.GetDBString(), []byte("Super secure plz no hax"))
	if err != nil {
		log.Fatalf("Cannot set up session db due to error: %v", err)
	}
	return sessionStore
}

var sessionStore = makeStore()

func AuthMiddleWare(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		// TODO: Hook session information into postgres backend
		session, err := sessionStore.Get(r, "SESSION")
		if err != nil {
			return "Authentication Error", &utils.AuthError{OriginalError: err}
		}
		if operationID == "PostLogin" {
			// check against db
			username, password, ok := r.BasicAuth()
			if !ok {
				return "Authentication Error", &utils.AuthError{NewError: errors.New("invalid basic auth header")}
			}
			stmt := SELECT(Users.Password).FROM(Users).WHERE(Users.Username.EQ(String(username)))
			var hashedPassword model.Users
			err := stmt.Query(utils.DB, &hashedPassword)
			if err != nil {
				if errors.Is(err, qrm.ErrNoRows) {
					return "Authentication Error", &utils.AuthError{OriginalError: errors.New("invalid username or password")}
				} else {
					return "Authentication Error", &utils.AuthError{OriginalError: err}
				}
			}
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.Password), []byte(password))
			if err != nil {
				if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
					return "Authentication Error", &utils.AuthError{OriginalError: errors.New("invalid username or password")}
				} else {
					return "Authentication Error", &utils.AuthError{OriginalError: err}
				}
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
			// Write cookie into session
			err = session.Save(r, w)
			if err != nil {
				return "Authentication Error", &utils.AuthError{OriginalError: err}
			}
		} else if operationID == "PostNewUser" {
			// TODO: INIT GAME AFTER NEW USER IS CREATED
			username, password, ok := r.BasicAuth()
			err = utils.CreateNewUser(username, password)
			if err != nil {
				return "Authentication Error", err
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
			// Write cookie into session
			err = session.Save(r, w)
			if err != nil {
				return "Authentication Error", &utils.AuthError{OriginalError: err}
			}
			if !ok {
				return "Authentication Error", &utils.AuthError{NewError: errors.New("invalid basic auth header")}
			}
		} else if session.IsNew {
			// We should have an existing session if we are not logging in or creating a new user
			return "Authentication Error", &utils.AuthError{NewError: errors.New("must use cookie to access this resource")}
		}
		ctx = context.WithValue(ctx, impl.USERNAME_VALUE, session.Values["username"])
		return f(ctx, w, r, request)
	}

}

func main() {
	fmt.Println("STARTING")
	RUN_TYPE := os.Args[1]

	if RUN_TYPE == "SERVER" {
		m := []nethttp.StrictHTTPMiddlewareFunc{AuthMiddleWare}
		server := impl.NewServer()
		// create a type that satisfies the `api.StrictServerInterface`, which contains an implementation of every operation from the generated code
		// i := api.NewStrictHandler(server, m)
		i := api.NewStrictHandlerWithOptions(server, m, api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				slog.Error("caught client error", "error", err.Error(), "code", http.StatusUnauthorized)
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				// All errors returned by nethttp.StrictHTTPMiddlewareFunc middleware are considered response errors,
				// even though bad auth is technically a bad request. (see codcgened stricthandler middleware for details)
				// We're going to be a little messy here and return client error codes with the response error options
				// so we don't have to rewrite the auth middleware.
				var authErr *utils.AuthError
				if errors.As(err, &authErr) {
					slog.Error("caught client error", "error", err.(*utils.AuthError).BothErrors(), "code", http.StatusUnauthorized)
					http.Error(w, err.Error(), http.StatusUnauthorized)
				} else {
					slog.Error("caught server error", "error", err.Error(), "code", http.StatusInternalServerError)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
		})
		r := http.NewServeMux()
		// get an `http.Handler` that we can use
		h := api.HandlerFromMux(i, r)

		s := &http.Server{
			Handler: h,
			Addr:    "127.0.0.1:8080",
		}
		// And we serve HTTP until the world ends.
		log.Fatal(s.ListenAndServe())
	} else if RUN_TYPE == "SCHEMA" {
		file, err := os.ReadFile("./schemas/schemas.sql")
		if err != nil {
			log.Fatal(err)
		}
		_, err = utils.OpenDB().Exec(string(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
