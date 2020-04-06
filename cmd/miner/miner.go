package main

import (
	"log"
	"net"

	"github.com/omzmarlon/blockfs/pkg/blockchain"
	rpc "github.com/omzmarlon/blockfs/pkg/grpc"
	"github.com/omzmarlon/blockfs/pkg/queue"
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

	// start internal services...
	blockchain := blockchain.New(blockchain.Conf{
		GenesisHash:    "abc",
		MinerID:        "123",
		OpDiffculty:    2,
		NoopDifficulty: 4,
	})
	queue.StartQueueDaemons(queue.Conf{
		MaxOpsQueueSize:        5,
		RetryBlockQueueMaxSize: 100,
		WaitOpsTimeoutMilli:    2000,
		BlockQueueSize:         100,
	}, blockchain)

	// Creates a new gRPC server
	log.Println("Starting BlockFS rpc service... ")
	s := grpc.NewServer()
	rpc.RegisterBlockFS(s)
	s.Serve(lis)

}
