package main

import (
	"log"
	"net"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"

	"github.com/anvitha0403/golog/internal/server"
	"github.com/anvitha0403/golog/internal/server/kvstore"
	"google.golang.org/grpc"
)

func main() {
	srv, err := server.NewHTTPServer(":8080")
	if err != nil {
		panic("http server error during initialization")
	}
	// gRPC server
	grpcServer := grpc.NewServer()
	store, err := kvstore.New(".")
	if err != nil {
		panic("grpc server error during initialization")
	}
	pb.RegisterKVStoreServiceServer(grpcServer, store)
	go func() {
		lis, _ := net.Listen("tcp", ":50051")
		grpcServer.Serve(lis)
	}()
	log.Fatal(srv.ListenAndServe())

}
