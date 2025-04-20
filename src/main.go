package main

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ryan-michael-19/web-drones/api"
	"github.com/ryan-michael-19/web-drones/impl"
	"github.com/ryan-michael-19/web-drones/utils/stateful"
	"github.com/ryan-michael-19/web-drones/utils/stateless"

	"github.com/antonlindstrom/pgstore"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/ryan-michael-19/web-drones/webdrones/public/model"
	. "github.com/ryan-michael-19/web-drones/webdrones/public/table"

	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"golang.org/x/crypto/bcrypt"
)

func getSessionEncryptionKey() []byte {
	var sessionEncryptionKey []byte
	sessionFile, present := os.LookupEnv("SESSION_KEY_FILE")
	if !present || sessionFile == "" { // session key file does not exist
		sessionEncrpytionKeyString, present := os.LookupEnv("SESSION_KEY")
		if !present || sessionEncrpytionKeyString == "" {
			slog.Warn("SESSION_KEY and SESSION_KEY_FILE are not set! Sessions will be encrypted with a hardcoded key.")
			sessionEncryptionKey = []byte("Super secure pls no hax")
		} else { // Session key variable exists
			sessionEncryptionKey = []byte(sessionEncrpytionKeyString)
		}
	} else { // Session key file exists
		var err error
		sessionEncryptionKey, err = os.ReadFile(sessionFile)
		if err != nil {
			log.Fatalf("Error reading session key file: %v", err)
		}
	}
	return sessionEncryptionKey
}

func makeStore() *pgstore.PGStore {
	// so we can use time objects in session.Values
	// https://stackoverflow.com/questions/24834480/using-custom-types-with-gorilla-sessions
	gob.Register(time.Time{})
	sessionEncryptionKey := getSessionEncryptionKey()
	sessionStore, err := pgstore.NewPGStore(stateful.GetDBString(), sessionEncryptionKey)
	if err != nil {
		log.Fatalf("Cannot set up session db due to error: %v", err)
	}
	return sessionStore
}

var sessionStore = makeStore()

func requestsPerSecondToTimeout(requestRate float64) float64 {
	return 1 / requestRate
}

var rateLimitLength = requestsPerSecondToTimeout(2)

type RateLimitError struct {
	message string
}

func (e *RateLimitError) Error() string {
	return e.message
}

func AuthMiddleWare(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		session, err := sessionStore.Get(r, "SESSION")
		if err != nil {
			return "Authentication Error", &stateless.AuthError{OriginalError: err}
		}

		lastRequest, ok := session.Values["lastRequest"]
		if ok && time.Since(lastRequest.(time.Time)).Seconds() < rateLimitLength {
			return "Timeout Error", &RateLimitError{message: "Rate limit reached. Please try again later."}
		} else { // last request does not exist or happened longer ago than the rate limit length
			session.Values["lastRequest"] = time.Now()
		}

		if operationID == "PostLogin" {
			// check against db
			username, password, ok := r.BasicAuth()
			if !ok {
				return "Authentication Error", &stateless.AuthError{NewError: errors.New("invalid basic auth header")}
			}
			stmt := SELECT(Users.Password).FROM(Users).WHERE(Users.Username.EQ(String(username)))
			var hashedPassword model.Users
			err := stmt.Query(stateful.DB, &hashedPassword)
			if err != nil {
				if errors.Is(err, qrm.ErrNoRows) {
					return "Authentication Error", &stateless.AuthError{OriginalError: errors.New("invalid username or password")}
				} else {
					return "Authentication Error", &stateless.AuthError{OriginalError: err}
				}
			}
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.Password), []byte(password))
			if err != nil {
				if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
					return "Authentication Error", &stateless.AuthError{OriginalError: errors.New("invalid username or password")}
				} else {
					return "Authentication Error", &stateless.AuthError{OriginalError: err}
				}
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
		} else if operationID == "PostNewUser" {
			// TODO: INIT GAME AFTER NEW USER IS CREATED
			username, password, ok := r.BasicAuth()
			err = stateful.CreateNewUser(username, password)
			if err != nil {
				return "Authentication Error", err
			}
			w.Header().Add("Content-Type", "text/plain")
			session.Values["username"] = username
			if !ok {
				return "Authentication Error", &stateless.AuthError{NewError: errors.New("invalid basic auth header")}
			}
		} else if session.IsNew {
			// We should have an existing session if we are not logging in or creating a new user
			return "Authentication Error", &stateless.AuthError{NewError: errors.New("must use cookie to access this resource")}
		}
		err = session.Save(r, w)
		if err != nil {
			return "Authentication Error", &stateless.AuthError{OriginalError: err}
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
				var authErr *stateless.AuthError
				var rateErr *RateLimitError
				if errors.As(err, &authErr) {
					slog.Error("caught client error", "error", err.(*stateless.AuthError).BothErrors(), "code", http.StatusUnauthorized)
					http.Error(w, err.Error(), http.StatusUnauthorized)
				} else if errors.As(err, &rateErr) {
					slog.Error("caught client error", "error", err.(*RateLimitError).Error(), "code", http.StatusUnauthorized)
					http.Error(w, err.Error(), http.StatusTooManyRequests)
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
			Addr:    "0.0.0.0:8080",
		}
		// And we serve HTTP until the world ends.
		log.Fatal(s.ListenAndServe())
	} else if RUN_TYPE == "SCHEMA" {
		file, err := os.ReadFile("./schemas/schemas.sql")
		if err != nil {
			log.Fatal(err)
		}
		_, err = stateful.OpenDB().Exec(string(file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
