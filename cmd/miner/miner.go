package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/config"
	"github.com/omzmarlon/blockfs/pkg/coordinator"
	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/flooder"
	rpc "github.com/omzmarlon/blockfs/pkg/grpc"
	grpc "google.golang.org/grpc"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "configs/miner1.json", "config file for miner")
	flag.Parse()

	minerConfig, err := config.ParseMinerConfig(configPath)
	if err != nil {
		log.Printf("Loading miner config failed with err %s", err)
		os.Exit(1)
	}
	log.Printf("Using configs at: %s with values:", configPath)
	spew.Dump(minerConfig)

	// communication channels among processes
	blocksProcessing := make(chan *domain.Block, minerConfig.BlocksProcessingBufferSize)
	opsProcessing := make(chan *domain.Op, minerConfig.OpsProcessingBufferSize)
	blocksFlooding := make(chan *domain.Block, minerConfig.BlockFloodingBufferSize)
	opsFlooding := make(chan *domain.Op, minerConfig.OpFloodingBufferSize)

	// start internal processes
	bc := blockchain.New(blockchain.Conf{
		GenesisHash:    minerConfig.BlockchainConfig.GenesisHash,
		MinerID:        minerConfig.MinerID,
		OpDiffculty:    minerConfig.BlockchainConfig.OpDifficulty,
		NoopDifficulty: minerConfig.BlockchainConfig.NoopDifficulty,
	})

	flooder := flooder.New(flooder.Conf{
		MinerID:                   minerConfig.MinerID,
		Port:                      minerConfig.FlooderConfig.PeerPort,
		PeerOpsFloodBufferSize:    minerConfig.FlooderConfig.PeerOpsFloodBufferSize,
		PeerBlocksFloodBufferSize: minerConfig.FlooderConfig.PeerBlocksFloodBufferSize,
		HeatbeatRetryMax:          minerConfig.FlooderConfig.HeatbeatRetryMax,
		Peers:                     config.ConvertPeerConfigsToMap(minerConfig.FlooderConfig.Peers),
	}, bc)
	flooder.StartFlooderDaemons(blocksProcessing, opsProcessing, blocksFlooding, opsFlooding)

	coordinator := coordinator.New(coordinator.Conf{
		MaxOpsQueueSize:        minerConfig.CoordinatorConfig.MaxOpsQueueSize,
		RetryBlockQueueMaxSize: minerConfig.CoordinatorConfig.RetryBlockQueueMaxSize,
		WaitOpsTimeoutMilli:    minerConfig.CoordinatorConfig.WaitOpsTimeoutMilli,
	}, bc)
	coordinator.StartCoordinatorDaemons(blocksProcessing, opsProcessing, blocksFlooding)

	// Creates a new gRPC server
	log.Println("Starting miner... ")
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(minerConfig.BlockfsPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		os.Exit(1)
	}
	log.Println("Starting BlockFS rpc service... ")
	s := grpc.NewServer()
	rpc.RegisterBlockFS(s)
	s.Serve(lis)
}
