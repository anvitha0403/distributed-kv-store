package main

import (
	"log"
	"net"

	"github.com/anvitha0403/golog/internal/server/auth"
	"github.com/anvitha0403/golog/internal/server/kvstore"
)

func main() {
	store, err := kvstore.ConnectFileStore(".")
	config := &Config{Store: store, Authorizer: auth.Authorizer{}}
	srv, err := NewHTTPServer(":8080", config)
	if err != nil {
		panic("http server error during initialization")
	}
	// gRPC server
	grpcServer, err := NewGRPCServer(config)
	if err != nil {
		panic("grpc server error during initialization")
	}

	go func() {
		lis, _ := net.Listen("tcp", ":50051")
		grpcServer.Serve(lis)
	}()
	log.Fatal(srv.ListenAndServe())

}
