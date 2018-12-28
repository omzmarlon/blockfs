package main

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"cs416/P1-v3d0b-q4d0b/bclib/chain"
	"cs416/P1-v3d0b-q4d0b/bclib/chain/generation"
	"cs416/P1-v3d0b-q4d0b/bclib/flooding"
	"cs416/P1-v3d0b-q4d0b/bclib/op"
	"cs416/P1-v3d0b-q4d0b/bclib/reqhandler"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	configPath := ""

	if len(os.Args)-1 == 1 {
		configPath = os.Args[1]
	} else {
		configPath = "./configs/miner1.json"
	}

	rawConfig, fileerr := ioutil.ReadFile(configPath)
	if fileerr != nil {
		log.Printf("Miner.go: Invalid json file: %s\n", fileerr)
		os.Exit(1)
	}
	minerConfig := &MinerConfig{}
	jsonErr := json.Unmarshal(rawConfig, minerConfig)
	if jsonErr != nil {
		log.Printf("Miner.go: Invalid json file: %s\n", jsonErr)
		os.Exit(1)
	}

	// Initializing blockchain
	bcmNewBlockInputChan := make(chan bclib.Block, 30)
	bcmInstance := chain.InitBlockChainManager(
		minerConfig.MinerID,
		minerConfig.GenesisBlockHash,
		int(minerConfig.MinedCoinsPerOpBlock),
		int(minerConfig.MinedCoinsPerNoOpBlock),
		int(minerConfig.NumCoinsPerFileCreate),
		int(minerConfig.ConfirmsPerFileCreate),
		int(minerConfig.ConfirmsPerFileAppend),
		int(minerConfig.PowPerOpBlock),
		int(minerConfig.PowPerNoOpBlock),
		bcmNewBlockInputChan)

	// Initialize OPs queue
	qml := uint32(3)
	opQueueNewOpInputChan, opQueueOuputChan := op.InitOpsQueue(minerConfig.GenOpBlockTimeout, qml)

	// Initialize flooder
	flooderConfig := flooding.FlooderConfig{
		MinerID:            minerConfig.MinerID,
		IncomingMinersAddr: minerConfig.IncomingMinersAddr,
		OutgoingMinersIP:   minerConfig.OutgoingMinersIP,
		PeerMinersAddrs:    minerConfig.PeerMinersAddrs,
		FDResponseIPPort:   minerConfig.OutgoingMinersIP + ":" + flooding.GetNewUnusedPort(),
		LostMsgThresh:      uint8(50),
	}
	flooderInstance, flderr := flooding.InitFloodingProtocol(flooderConfig, opQueueNewOpInputChan, bcmNewBlockInputChan, bcmInstance)
	if flderr != nil {
		log.Printf("Miner.go: could not initialize flooder due to error: %s\n", flderr)
		os.Exit(1)
	}

	// Initialize block generator
	generation.InitBlockGenerator(
		minerConfig.MinerID,
		minerConfig.GenesisBlockHash,
		minerConfig.PowPerOpBlock,
		minerConfig.PowPerNoOpBlock,
		minerConfig.ConfirmsPerFileCreate,
		minerConfig.ConfirmsPerFileAppend,
		bcmNewBlockInputChan,
		bcmInstance,
		opQueueOuputChan,
		flooderInstance)

	// Initialize Request Handler
	reqerr := reqhandler.InitReqHandler(
		minerConfig.IncomingClientsAddr,
		minerConfig.MinerID,
		int(minerConfig.ConfirmsPerFileCreate),
		int(minerConfig.ConfirmsPerFileAppend),
		int(minerConfig.MinedCoinsPerOpBlock),
		int(minerConfig.MinedCoinsPerNoOpBlock),
		int(minerConfig.NumCoinsPerFileCreate),
		flooderInstance,
		bcmInstance)

	if reqerr != nil {
		log.Printf("Miner.go: could not initialize request handler due to error: %s\n", reqerr)
	}

	log.Printf("Miner.go: successfully initialized miner: %s\n", minerConfig.MinerID)
	<-bclib.Done
}

// MinerConfig holds all config variables for a miner
type MinerConfig struct {
	MinedCoinsPerOpBlock   uint8
	MinedCoinsPerNoOpBlock uint8
	NumCoinsPerFileCreate  uint8
	GenOpBlockTimeout      uint8
	GenesisBlockHash       string
	PowPerOpBlock          uint8
	PowPerNoOpBlock        uint8
	ConfirmsPerFileCreate  uint8
	ConfirmsPerFileAppend  uint8
	MinerID                string
	PeerMinersAddrs        []string
	IncomingMinersAddr     string
	OutgoingMinersIP       string
	IncomingClientsAddr    string
}
