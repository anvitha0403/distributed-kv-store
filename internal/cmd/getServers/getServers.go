package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	api "github.com/anvitha0403/golog/internal/server/api/v1"
	"github.com/anvitha0403/golog/internal/server/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	addr := flag.String("addr", ":8400", "service address")
	flag.Parse()

	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: *addr,
	})
	if err != nil {
		panic(err)
	}
	tlsCreds := credentials.NewTLS(peerTLSConfig)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}

	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		panic(err)
	}
	client := api.NewKVStoreServiceClient(conn)
	ctx := context.Background()
	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("servers:")
	for _, server := range res.Servers {
		fmt.Printf("\t- %v\n", server)
	}

	putResponse, err := client.Put(ctx, &api.KVPair{Key: "name", Value: "lalitha"})
	fmt.Println(putResponse, err)

	getResponse, err := client.Get(ctx, &api.KVPair{Key: "name", Value: "lalitha"})
	fmt.Println(getResponse, err)

	delResponse, err := client.Delete(ctx, &api.KVPair{Key: "name", Value: "lalitha"})
	fmt.Println(delResponse, err)
	getResponse, err = client.Get(ctx, &api.KVPair{Key: "name", Value: "lalitha"})
	fmt.Println(getResponse, err)

}
