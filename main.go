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

	. "colony-bots/webdrones/public/table"

	. "github.com/go-jet/jet/v2/postgres"

	"github.com/gorilla/sessions"
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
						// records, err := schemas.OpenDB(ctx).Query(ctx,
						// "SELECT password FROM users WHERE username = $1",
						// username,
						// )
						stmt := SELECT(Users.Password).FROM(Users).WHERE(Users.Username.EQ(String(username)))
						var hashedPassword string
						err := stmt.Query(schemas.OpenDB(), &hashedPassword)
						if err != nil {
							return "Authentication Error", err
						}
						err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
						if err != nil {
							// TODO: Return 401
							return "Authentication Error", err
						}
						session.Values["username"] = username
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
						stmt := Users.INSERT(
							Users.CreatedAt, Users.UpdatedAt, Users.Username, Users.Password,
						).VALUES(
							NOW(), NOW(), String(username), String(string(hashedPassword)),
						)
						_, err = stmt.Exec(schemas.OpenDB())
						if err != nil {
							// TODO: Return 401
							return "Authentication Error", err
						}
						session.Values["username"] = username
					}
					w.Header().Add("Content-Type", "text/plain")
					// Always set up cookies/sessions no matter what kind of request we've sent
					err = session.Save(r, w)
					if err != nil {
						return "Authentication Error", err
					}
					ctx = context.WithValue(ctx, impl.USERNAME_VALUE, session.Values["username"])
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
		_, err = schemas.OpenDB().Exec(string(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
