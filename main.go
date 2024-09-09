package main

import (
	"colony-bots/api"
	"colony-bots/impl"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
)

func main() {
	RUN_TYPE := os.Args[1]

	if RUN_TYPE == "SERVER" {
		// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
		server := impl.NewServer()

		r := http.NewServeMux()

		// get an `http.Handler` that we can use
		h := api.HandlerFromMux(server, r)

		s := &http.Server{
			Handler: h,
			Addr:    "0.0.0.0:8080",
		}

		// And we serve HTTP until the world ends.
		log.Fatal(s.ListenAndServe())
	} else if RUN_TYPE == "C" || RUN_TYPE == "CLIENT" {
		hc := http.Client{}
		client, err := api.NewClient("http://localhost:8080", api.WithHTTPClient(&hc))
		if err != nil {
			log.Fatal(err)
		}
		// learning the hard way why go devs advise against reflection !!
		t := reflect.TypeOf(client)
		clientMethods := make([]reflect.Method, t.NumMethod())
		for i := 0; i < t.NumMethod(); i++ {
			clientMethods[i] = t.Method(i)
		}
		fn := os.Args[2]
		args := make([]reflect.Value, len(os.Args[3:])+2)
		args[0] = reflect.ValueOf(client)
		args[1] = reflect.ValueOf(context.TODO()) // TODO: This works apparently??
		for i := range os.Args[3:] {
			args[i+1] = reflect.ValueOf(os.Args[3+i])
		}
		for _, m := range clientMethods {
			if m.Name == fn {
				return_value := m.Func.Call(args)
				response := return_value[0].Elem()
				if err := return_value[1].Interface(); err != nil {
					log.Fatal(err)
				}
				contentLength := response.FieldByName("ContentLength").Int()
				resBody := make([]byte, contentLength)
				response.FieldByName("Body").Interface().(io.ReadCloser).Read(resBody)
				fmt.Println(string(resBody))
			}
		}

	}
}
