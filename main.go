package main

import (
	"colony-bots/api"
	"colony-bots/impl"
	"colony-bots/schemas"
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	RUN_TYPE := os.Args[1]

	if RUN_TYPE == "SERVER" {
		sessionStore := sessions.NewCookieStore([]byte("Super secure plz no hax")) // TODO: SET UP ENCRYPTION KEYS
		m := []nethttp.StrictHTTPMiddlewareFunc{
			func(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
				return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
					session, err := sessionStore.Get(r, "SESSION")
					if err != nil {
						return "Authentication Error", err
					}
					if operationID == "PostLogin" {
						// check against db
						username, password, ok := r.BasicAuth()
						if !ok {
							// TODO: Return 401
							return "Authentication Error", errors.New("Basic Auth Header issue")
						}
						records, err := schemas.OpenDB(ctx).Query(ctx,
							"SELECT password FROM users WHERE username = $1",
							username,
						)
						if err != nil {
							return "Authentication Error", err
						}
						// TODO: create struct when I'm being less lazy
						hashedPassword, err := pgx.CollectExactlyOneRow(records, pgx.RowTo[string])
						if err != nil {
							return "Authentication Error", err
						}
						records.Close()
						err = records.Err()
						if err != nil {
							return "Authentication error", err
						}
						err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
						if err != nil {
							// TODO: Return 401
							return "Authentication Error", err
						}
						ctx = context.WithValue(ctx, impl.USERNAME_VALUE, username)
					} else if operationID == "PostNewUser" {
						// add new user to db
						username, password, ok := r.BasicAuth()
						if !ok {
							// TODO: Return 401
							return "Authentication Error", errors.New("Basic Auth Header issue")
						}
						hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
						if err != nil {
							return "Authentication Error", err
						}
						stmt := schemas.BuildInsert(
							"users", "username", "password",
						)
						_, err = schemas.OpenDB(ctx).Exec(ctx, stmt, username, string(hashedPassword))
						if err != nil {
							// TODO: Return 401
							return "Authentication Error", err
						}
						ctx = context.WithValue(ctx, impl.USERNAME_VALUE, username)
					}
					w.Header().Add("Content-Type", "text/plain")
					// Always set up cookies/sessions no matter what kind of request we've sent
					err = session.Save(r, w)
					if err != nil {
						return "Authentication Error", err
					}
					ctx = context.WithValue(ctx, impl.SESSION_VALUE, w.Header().Get("Set-Cookie"))
					return f(ctx, w, r, request)
				}
			},
		}
		server := impl.NewServer()
		// create a type that satisfies the `api.StrictServerInterface`, which contains an implementation of every operation from the generated code
		i := api.NewStrictHandler(server, m)

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
		ctx := context.Background()
		_, err = schemas.OpenDB(ctx).Exec(ctx, string(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
