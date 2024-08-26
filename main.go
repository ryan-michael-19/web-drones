package main

import (
	"colony-bots/api"
	"colony-bots/impl"
	"colony-bots/schemas"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	} else if RUN_TYPE == "SCHEMA" {
		dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=EST"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		db.AutoMigrate(&schemas.Bots{})
		db.AutoMigrate(&schemas.Mines{})
		db.AutoMigrate(&schemas.BotActions{})
	}
}
