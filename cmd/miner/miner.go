package main

import (
	"log"
	"net"

	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/coordinator"
	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/flooder"
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

	// config values
	minerID := "123"
	blocksProcessingBufferSize := 100
	opsProcessingBufferSize := 100
	opFloodingBufferSize := 100
	blockFloodingBufferSize := 100

	// communication channels among processes
	blocksProcessing := make(chan *domain.Block, blocksProcessingBufferSize)
	opsProcessing := make(chan *domain.Op, opsProcessingBufferSize)
	blocksFlooding := make(chan *domain.Block, blockFloodingBufferSize)
	opsFlooding := make(chan *domain.Op, opFloodingBufferSize)

	// start internal processes
	bc := blockchain.New(blockchain.Conf{
		GenesisHash:    "abc",
		MinerID:        minerID,
		OpDiffculty:    2,
		NoopDifficulty: 4,
	})

	flooder := flooder.New(flooder.Conf{
		MinerID:                   minerID,
		Port:                      8080,
		PeerOpsFloodBufferSize:    100,
		PeerBlocksFloodBufferSize: 100,
		HeatbeatRetryMax:          3,
		Peers:                     make(map[string]string),
	}, bc)
	flooder.StartFlooderDaemons(blocksProcessing, opsProcessing, blocksFlooding, opsFlooding)

	coordinator := coordinator.New(coordinator.Conf{
		MaxOpsQueueSize:        5,
		RetryBlockQueueMaxSize: 100,
		WaitOpsTimeoutMilli:    1500,
	}, bc)
	coordinator.StartCoordinatorDaemons(blocksProcessing, opsProcessing, blocksFlooding)

	// Creates a new gRPC server
	log.Println("Starting BlockFS rpc service... ")
	s := grpc.NewServer()
	rpc.RegisterBlockFS(s)
	s.Serve(lis)
}
