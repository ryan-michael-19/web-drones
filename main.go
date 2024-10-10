package main

import (
	"colony-bots/api"
	"colony-bots/impl"
	"colony-bots/schemas"
	"context"
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
						log.Fatal(err)
					}
					if operationID == "Login" {
						// check against db
						username, password, ok := r.BasicAuth()
						if !ok {
							// TODO: Return 401
							log.Fatal("Basic auth issue")
						}
						records, err := schemas.OpenDB(ctx).Query(ctx,
							"SELECT password FROM users WHERE username = $1",
							username,
						)
						if err != nil {
							log.Fatal(err)
						}
						// TODO: create struct when I'm being less lazy
						hashedPassword, err := pgx.CollectExactlyOneRow(records, pgx.RowTo[string])
						if err != nil {
							log.Fatal(err)
						}
						err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
						if err != nil {
							// TODO: Return 401
							log.Fatal("Invalid Password")
						}
						session.Values["username"] = username
					} else if operationID == "NewUser" {
						// add new user to db
						username, password, ok := r.BasicAuth()
						if !ok {
							// TODO: Return 401
							log.Fatal("Basic auth issue")
						}
						hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
						if err != nil {
							log.Fatal(err)
						}
						stmt := schemas.BuildInsert(
							"users", "username", "password",
						)
						_, err = schemas.OpenDB(ctx).Exec(ctx, stmt, username, string(hashedPassword))
						if err != nil {
							// TODO: Return 401
							log.Fatal(err)
						}
						session.Values["username"] = username
					}
					// Always set up cookies/sessions no matter what kind of request we've sent
					err = session.Save(r, w)
					if err != nil {
						log.Fatal(err)
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
