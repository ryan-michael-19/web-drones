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
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

func main() {
	RUN_TYPE := os.Args[1]

	if RUN_TYPE == "SERVER" {
		sessionStore := sessions.NewCookieStore() // TODO: Does this need a key?
		m := []nethttp.StrictHTTPMiddlewareFunc{
			func(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
				return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
					sessionStore.Get(r, "session")
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
	// removing this and making a TS client that will be much easier to use
	// } else if RUN_TYPE == "C" || RUN_TYPE == "CLIENT" {
	// hc := http.Client{}
	// client, err := api.NewClient("http://localhost:8080", api.WithHTTPClient(&hc))
	// if err != nil {
	// log.Fatal(err)
	// }
	// // learning the hard way why go devs advise against reflection !!
	// t := reflect.TypeOf(client)
	// clientMethods := make([]reflect.Method, t.NumMethod())
	// for i := 0; i < t.NumMethod(); i++ {
	// clientMethods[i] = t.Method(i)
	// }
	// fn := os.Args[2]
	// args := make([]reflect.Value, len(os.Args[3:])+2)
	// args[0] = reflect.ValueOf(client)
	// args[1] = reflect.ValueOf(context.Background()) // TODO: This works apparently??
	// for i := range os.Args[3:] {
	// args[i+1] = reflect.ValueOf(os.Args[3+i])
	// }
	// for _, m := range clientMethods {
	// if m.Name == fn {
	// return_value := m.Func.Call(args)
	// response := return_value[0].Elem()
	// if err := return_value[1].Interface(); err != nil {
	// log.Fatal(err)
	// }
	// contentLength := response.FieldByName("ContentLength").Int()
	// resBody := make([]byte, contentLength)
	// response.FieldByName("Body").Interface().(io.ReadCloser).Read(resBody)
	// fmt.Println(string(resBody))
	// }
	// }
	//
	// }
}
