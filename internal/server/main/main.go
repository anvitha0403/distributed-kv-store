package main

import (
	"log"

	"github.com/anvitha0403/golog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())

}
