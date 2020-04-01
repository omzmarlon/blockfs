package main

import (
	"log"
	"net"

	rpc "github.com/omzmarlon/blockfs/pkg/grpc"
	grpc "google.golang.org/grpc"
)

// TODO
// - port should be configurable

const (
	port = ":50051"
)

func main() {
	log.Println("Starting miner... ")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Creates a new gRPC server
	log.Println("Starting BlockFS rpc service... ")
	s := grpc.NewServer()
	rpc.RegisterBlockFS(s)
	s.Serve(lis)

}
